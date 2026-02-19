package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
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

func (r *RedisNewsRepository) CreateNews(news *pb.News) error {
    ctx := context.Background()
    
    // Generate ID directly here to use as key
    id := uuid.New().String()
    
    // 1. Store News object as JSON
    data, err := json.Marshal(news)
    if err != nil {
        return err
    }
    
    if err := r.client.Set(ctx, "news:"+id, data, 0).Err(); err != nil {
        return err
    }

    // 2. Add to Sorted Set 'news:latest' with timestamp as score
    // Using current time as score for sorting
    score := float64(time.Now().Unix())
    
    if err := r.client.ZAdd(ctx, "news:latest", redis.Z{
        Score:  score,
        Member: id,
    }).Err(); err != nil {
        return err
    }

    // 3. Add to City specific Sorted Set
    if news.City != "" {
        if err := r.client.ZAdd(ctx, "news:city:"+news.City, redis.Z{
            Score:  score,
            Member: id,
        }).Err(); err != nil {
            return err
        }
    }
    
    fmt.Printf("News created in Redis: %s (%s)\n", news.Name, id)
    return nil
}

func (r *RedisNewsRepository) GetLatestNews(limit int) ([]*pb.News, error) {
    ctx := context.Background()
    
    // Get latest IDs
    ids, err := r.client.ZRevRange(ctx, "news:latest", 0, int64(limit-1)).Result()
    if err != nil {
        return nil, err
    }
    
    return r.fetchNewsByIDs(ctx, ids)
}

func (r *RedisNewsRepository) GetLatestNewsInCity(city string, limit int) ([]*pb.News, error) {
    ctx := context.Background()
    
    // Get latest IDs for city
    ids, err := r.client.ZRevRange(ctx, "news:city:"+city, 0, int64(limit-1)).Result()
    if err != nil {
        return nil, err
    }
    
    return r.fetchNewsByIDs(ctx, ids)
}

func (r *RedisNewsRepository) fetchNewsByIDs(ctx context.Context, ids []string) ([]*pb.News, error) {
    if len(ids) == 0 {
        return []*pb.News{}, nil
    }

    // Prepare keys
    keys := make([]string, len(ids))
    for i, id := range ids {
        keys[i] = "news:" + id
    }

    // MGet to fetch all news items efficiently
    vals, err := r.client.MGet(ctx, keys...).Result()
    if err != nil {
        return nil, err
    }

    var newsList []*pb.News
    for _, val := range vals {
        if val == nil {
            continue
        }
        strVal, ok := val.(string)
        if !ok {
            continue
        }
        
        var n pb.News
        if err := json.Unmarshal([]byte(strVal), &n); err != nil {
            continue
        }
        newsList = append(newsList, &n)
    }
    
    return newsList, nil
}
