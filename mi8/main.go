package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "os"

    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    "google.golang.org/protobuf/types/known/emptypb"
    pb "mi8/proto"
)

type server struct {
    pb.UnimplementedMI8ServiceServer
    repo NewsRepository
}

func (s *server) GetLatestNews(ctx context.Context, in *pb.GetLatestNewsRequest) (*pb.NewsList, error) {
    news, err := s.repo.GetLatestNews(int(in.Limit))
    if err != nil {
        return nil, err
    }
    return &pb.NewsList{News: news}, nil
}

func (s *server) GetLatestNewsInCity(ctx context.Context, in *pb.GetLatestNewsInCityRequest) (*pb.NewsList, error) {
    news, err := s.repo.GetLatestNewsInCity(in.City, int(in.Limit))
    if err != nil {
        return nil, err
    }
    return &pb.NewsList{News: news}, nil
}

func (s *server) CreateNews(ctx context.Context, in *pb.News) (*emptypb.Empty, error) {
    err := s.repo.CreateNews(in)
    if err != nil {
        return nil, err
    }
    return &emptypb.Empty{}, nil
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    s := grpc.NewServer()
    
    // Initialize Repository (Redis)
    var repo NewsRepository
    redisAddr := os.Getenv("REDIS_ADDR")
    if redisAddr != "" {
        fmt.Println("Using Redis Repository at " + redisAddr)
        r, err := NewRedisNewsRepository()
        if err != nil {
             log.Fatalf("failed to create redis repo: %v", err)
        }
        repo = r
    } else {
        fmt.Println("Using In-Memory Repository (Default)")
        repo = NewArrayNewsRepository()
    }
    
    pb.RegisterMI8ServiceServer(s, &server{repo: repo})

    // Register reflection service often useful for tools like grpcurl
    reflection.Register(s)

    fmt.Println("MI8 gRPC Server listening at :50051")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
