package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var collection *mongo.Collection

func main() {
	// 1. Connect to MongoDB
	mongoURI := "mongodb://localhost:27017"
	if os.Getenv("MONGODB_URI") != "" {
		mongoURI = os.Getenv("MONGODB_URI")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	// Ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
    
    collection = client.Database("erasmumu").Collection("offers")
	fmt.Println("Connected to MongoDB!")

	// 2. Setup Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Routes
	r.Post("/offer", createOffer)
	r.Get("/offer/{id}", getOffer)
	r.Get("/offer", getOffers) // Handles domain and city filters
	r.Put("/offer/{id}", updateOffer)
	r.Delete("/offer/{id}", deleteOffer)

	fmt.Println("Erasmumu Service starting on :8080")
	http.ListenAndServe(":8080", r)
}
