package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
    
    // Initialize Repository
    repo := NewArrayNewsRepository()
    
	pb.RegisterMI8ServiceServer(s, &server{repo: repo})

	// Register reflection service often useful for tools like grpcurl
	reflection.Register(s)

	fmt.Println("MI8 gRPC Server listening at :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
