package main

import (
	"context"
	"encoding/json"

	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type Student struct {
	ID        int    `json:"id"`
	Firstname string `json:"firstname"`
	Name      string `json:"name"`
	Domain    string `json:"domain"`
}

// Global DB connection
var db *pgx.Conn

// Helper for JSON responses
func jsonResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// POST /student
func createStudent(w http.ResponseWriter, r *http.Request) {
	var s Student
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := db.QueryRow(context.Background(),
		"INSERT INTO students (firstname, name, domain) VALUES ($1, $2, $3) RETURNING id",
		s.Firstname, s.Name, s.Domain).Scan(&s.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusCreated, s)
}

// GET /student/:id
func getStudent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var s Student
	err = db.QueryRow(context.Background(),
		"SELECT id, firstname, name, domain FROM students WHERE id=$1", id).Scan(&s.ID, &s.Firstname, &s.Name, &s.Domain)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Student not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	jsonResponse(w, http.StatusOK, s)
}

// GET /student?domain=<domain>
func getStudents(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	
	query := "SELECT id, firstname, name, domain FROM students"
	args := []interface{}{}
	if domain != "" {
		query += " WHERE domain=$1"
		args = append(args, domain)
	}

	rows, err := db.Query(context.Background(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var students []Student
	for rows.Next() {
		var s Student
		if err := rows.Scan(&s.ID, &s.Firstname, &s.Name, &s.Domain); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		students = append(students, s)
	}

	jsonResponse(w, http.StatusOK, students)
}

// PUT /student/:id
func updateStudent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var s Student
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = db.Exec(context.Background(),
		"UPDATE students SET firstname=$1, name=$2, domain=$3 WHERE id=$4",
		s.Firstname, s.Name, s.Domain, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
    
    s.ID = id
	jsonResponse(w, http.StatusOK, s)
}

// DELETE /student/:id
func deleteStudent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(context.Background(), "DELETE FROM students WHERE id=$1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
