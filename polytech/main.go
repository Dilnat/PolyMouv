package main

import (
	"context"
	"fmt"

	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 1. Connect to Database
	dbUrl := "postgres://postgres:postgres@localhost:5432/polytech"
	// Override with env var if needed
	if os.Getenv("DATABASE_URL") != "" {
		dbUrl = os.Getenv("DATABASE_URL")
	}

	var err error
	db, err = pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create table if not exists
	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS students (
			id SERIAL PRIMARY KEY,
			firstname TEXT NOT NULL,
			name TEXT NOT NULL,
			domain TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS registrations (
			id SERIAL PRIMARY KEY,
			student_id INT NOT NULL,
			offer_id TEXT NOT NULL,
			status TEXT NOT NULL,
			message TEXT
		);
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create table: %v\n", err)
		os.Exit(1)
	}

	// 2. Setup Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/student", createStudent)
	r.Get("/student/{id}", getStudent) // chi uses {param} syntax
	r.Get("/student", getStudents)
	r.Put("/student/{id}", updateStudent)
	r.Delete("/student/{id}", deleteStudent)

	r.Post("/internship", registerInternship)
	r.Get("/internship/{id}", getRegistration)

	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", r)
}
