package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/redis/go-redis/v9"
    pb "mi8/proto"
)

type RedisNewsRepository struct {
    client *redis.Client
}

func NewRedisNewsRepository() (*RedisNewsRepository, error) {
    redisAddr := os.Getenv("REDIS_ADDR")
    if redisAddr == "" {
        redisAddr = "localhost:6379"
    }

    client := redis.NewClient(&redis.Options{
        Addr: redisAddr,
    })

    if err := client.Ping(context.Background()).Err(); err != nil {
        return nil, fmt.Errorf("failed to connect to redis: %v", err)
    }

    return &RedisNewsRepository{client: client}, nil
}

// Helper to calculate score impacts
func calculateImpact(tags []string) (int32, int32, int32, int32) {
    var dSafety, dEconomy, dQoL, dCulture int32
    
    for _, t := range tags {
        tag := strings.ToLower(t)
        switch tag {
        case "innovation":
            dSafety += 20; dEconomy += 60; dQoL += 30; dCulture += 5
        case "culture":
            dEconomy += 15; dQoL += 40; dCulture += 75
        case "healthcare":
            dSafety += 30; dEconomy += 20; dQoL += 30
        case "entertainment":
            dEconomy += 20; dQoL += 25; dCulture += 35
        case "crisis":
            dSafety -= 80; dEconomy -= 100; dQoL -= 60; dCulture -= 30
        case "crime":
            dSafety -= 120; dEconomy -= 50; dQoL -= 80; dCulture -= 40
        case "disaster":
            dSafety -= 100; dEconomy -= 70; dQoL -= 90; dCulture -= 30
        }
    }
    return dSafety, dEconomy, dQoL, dCulture
}

func (r *RedisNewsRepository) CreateNews(news *pb.News) error {
    ctx := context.Background()
    id := uuid.New().String()
    
    // 1. Store News
    data, err := json.Marshal(news)
    if err != nil { return err }
    if err := r.client.Set(ctx, "news:"+id, data, 0).Err(); err != nil { return err }

    score := float64(time.Now().Unix())
    r.client.ZAdd(ctx, "news:latest", redis.Z{Score: score, Member: id})
    if news.City != "" {
        r.client.ZAdd(ctx, "news:city:"+news.City, redis.Z{Score: score, Member: id})
        
        // 2. Update City Scores
        dSafety, dEconomy, dQoL, dCulture := calculateImpact(news.Tags)
        
        cityKey := "city:score:" + news.City
        
        // Initialize if not exists (HSETNX or check exits)
        exists, _ := r.client.Exists(ctx, cityKey).Result()
        if exists == 0 {
            // Default base scores
            r.client.HSet(ctx, cityKey, map[string]interface{}{
                "safety": 1000, "economy": 1000, "qol": 1000, "culture": 1000,
                "country": news.Country,
            })
        } else {
            // Ensure country is set if missing
             r.client.HSetNX(ctx, cityKey, "country", news.Country)
        }
        
        // Atomic increments
        // Note: checking < 0 in Redis Lua script would be atomic, but for simple task we just HINCRBY
        // and fixing to 0 is acceptable/simpler.
        
        pipe := r.client.Pipeline()
        pipe.HIncrBy(ctx, cityKey, "safety", int64(dSafety))
        pipe.HIncrBy(ctx, cityKey, "economy", int64(dEconomy))
        pipe.HIncrBy(ctx, cityKey, "qol", int64(dQoL))
        pipe.HIncrBy(ctx, cityKey, "culture", int64(dCulture))
        pipe.HSet(ctx, cityKey, "last_updated", news.Date)
        
        _, err := pipe.Exec(ctx)
        if err != nil {
            fmt.Printf("Error updating scores: %v\n", err)
        }
        
        // Update Ranking Sorted Set
        // Total score = Sum of 4 components? Or some average. Let's sum.
        vals, _ := r.client.HMGet(ctx, cityKey, "safety", "economy", "qol", "culture").Result()
        var total int64
        for _, v := range vals {
            if v != nil {
                i, _ := strconv.ParseInt(v.(string), 10, 64)
                if i < 0 { i = 0 } // Clamp to 0
                total += i
            }
        }
        
        // Also clamp values in Hash? Not strictly required by prompt but good practice.
        // For efficiency, just use total here.
        
        r.client.ZAdd(ctx, "cities:rank", redis.Z{
            Score: float64(total),
            Member: news.City,
        })
    }
    
    fmt.Printf("News created: %s, updated scores for %s\n", news.Name, news.City)
    return nil
}

func (r *RedisNewsRepository) GetLatestNews(limit int) ([]*pb.News, error) {
    ctx := context.Background()
    ids, err := r.client.ZRevRange(ctx, "news:latest", 0, int64(limit-1)).Result()
    if err != nil { return nil, err }
    return r.fetchNewsByIDs(ctx, ids)
}

func (r *RedisNewsRepository) GetLatestNewsInCity(city string, limit int) ([]*pb.News, error) {
    ctx := context.Background()
    ids, err := r.client.ZRevRange(ctx, "news:city:"+city, 0, int64(limit-1)).Result()
    if err != nil { return nil, err }
    return r.fetchNewsByIDs(ctx, ids)
}

func (r *RedisNewsRepository) GetCityScore(city string) (*pb.CityScore, error) {
    ctx := context.Background()
    cityKey := "city:score:" + city
    
    res, err := r.client.HGetAll(ctx, cityKey).Result()
    if err != nil { return nil, err }
    
    if len(res) == 0 {
        return nil, fmt.Errorf("city scores not found")
    }
    
    toInt := func(s string) int32 {
        v, _ := strconv.Atoi(s)
        if v < 0 { return 0 }
        return int32(v)
    }
    
    return &pb.CityScore{
        City: city,
        Country: res["country"],
        Safety: toInt(res["safety"]),
        Economy: toInt(res["economy"]),
        QualityOfLife: toInt(res["qol"]),
        Culture: toInt(res["culture"]),
        LastUpdated: res["last_updated"],
    }, nil
}

func (r *RedisNewsRepository) GetTopCities(limit int) ([]*pb.CityScore, error) {
    ctx := context.Background()
    
    cities, err := r.client.ZRevRange(ctx, "cities:rank", 0, int64(limit-1)).Result()
    if err != nil { return nil, err }
    
    var scores []*pb.CityScore
    for _, city := range cities {
        score, err := r.GetCityScore(city)
        if err == nil {
            scores = append(scores, score)
        }
    }
    return scores, nil
}


func (r *RedisNewsRepository) fetchNewsByIDs(ctx context.Context, ids []string) ([]*pb.News, error) {
    if len(ids) == 0 { return []*pb.News{}, nil }
    keys := make([]string, len(ids))
    for i, id := range ids { keys[i] = "news:" + id }
    vals, err := r.client.MGet(ctx, keys...).Result()
    if err != nil { return nil, err }

    var newsList []*pb.News
    for _, val := range vals {
        if val == nil { continue }
        strVal, ok := val.(string)
        if !ok { continue }
        var n pb.News
        if json.Unmarshal([]byte(strVal), &n) == nil {
            newsList = append(newsList, &n)
        }
    }
    return newsList, nil
}
