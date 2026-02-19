package main

import (
	pb "mi8/proto"
	"strings"
	"time"
    "fmt"
)

// NewsRepository interface
type NewsRepository interface {
	GetLatestNews(limit int) ([]*pb.News, error)
	GetLatestNewsInCity(city string, limit int) ([]*pb.News, error)
	CreateNews(news *pb.News) error
    GetCityScore(city string) (*pb.CityScore, error)
    GetTopCities(limit int) ([]*pb.CityScore, error)
}

// ArrayNewsRepository implementation
type ArrayNewsRepository struct {
	newsStore []*pb.News
}

func NewArrayNewsRepository() *ArrayNewsRepository {
	repo := &ArrayNewsRepository{
		newsStore: []*pb.News{},
	}
    // Populate with test data
    repo.CreateNews(&pb.News{
        Name:    "Tech Boom in Berlin",
        Source:  "TechDaily",
        Date:    time.Now().Format(time.RFC3339),
        Tags:    []string{"Tech", "Economy"},
        City:    "Berlin",
        Country: "Germany",
    })
    repo.CreateNews(&pb.News{
        Name:    "New Startup Hub in Paris",
        Source:  "LeMonde",
        Date:    time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
        Tags:    []string{"Startup", "Business"},
        City:    "Paris",
        Country: "France",
    })
    repo.CreateNews(&pb.News{
        Name:    "Another Tech Event",
        Source:  "TechDaily",
        Date:    time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
        Tags:    []string{"Tech"},
        City:    "Berlin",
        Country: "Germany",
    })
    repo.CreateNews(&pb.News{
        Name:    "AI Conference 2026",
        Source:  "AI News",
        Date:    time.Now().Format(time.RFC3339),
        Tags:    []string{"AI", "Tech"},
        City:    "San Francisco",
        Country: "USA",
    })
	return repo
}

func (r *ArrayNewsRepository) CreateNews(news *pb.News) error {
    // Prepend for "latest" behavior efficiently, or just append and sort on query.
    // Given the requirement "GetLatestNews", usually newest first.
    // Simple prepend here.
    r.newsStore = append([]*pb.News{news}, r.newsStore...)
    fmt.Printf("News created: %s in %s\n", news.Name, news.City)
	return nil
}

func (r *ArrayNewsRepository) GetLatestNews(limit int) ([]*pb.News, error) {
	if limit > len(r.newsStore) {
		limit = len(r.newsStore)
	}
    return r.newsStore[:limit], nil
}

func (r *ArrayNewsRepository) GetLatestNewsInCity(city string, limit int) ([]*pb.News, error) {
	var filtered []*pb.News
	for _, n := range r.newsStore {
		if strings.EqualFold(n.City, city) {
			filtered = append(filtered, n)
		}
	}
    if limit > len(filtered) {
        limit = len(filtered)
    }
	return filtered[:limit], nil
}

func (r *ArrayNewsRepository) GetCityScore(city string) (*pb.CityScore, error) {
    return &pb.CityScore{City: city, Safety: 1000}, nil
}

func (r *ArrayNewsRepository) GetTopCities(limit int) ([]*pb.CityScore, error) {
    return []*pb.CityScore{}, nil
}
