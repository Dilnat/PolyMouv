package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "polytech/proto"
)

type MI8Client struct {
	client pb.MI8ServiceClient
    conn *grpc.ClientConn
}

func NewMI8Client(addr string) (*MI8Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}
	c := pb.NewMI8ServiceClient(conn)
	return &MI8Client{client: c, conn: conn}, nil
}

func (c *MI8Client) Close() {
    c.conn.Close()
}

func (c *MI8Client) GetLatestNews(limit int32) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	r, err := c.client.GetLatestNews(ctx, &pb.GetLatestNewsRequest{Limit: limit})
	if err != nil {
		log.Printf("could not get news: %v", err)
        return
	}
	
    fmt.Printf("--- Latest News from MI8 (Limit: %d) ---\n", limit)
	for _, n := range r.GetNews() {
		fmt.Printf("Title: %s | City: %s | Date: %s\n", n.Name, n.City, n.Date)
	}
    fmt.Println("----------------------------------------")
}
