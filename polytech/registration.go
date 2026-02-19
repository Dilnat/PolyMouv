package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type Registration struct {
	ID        int    `json:"id"`
	StudentID int    `json:"studentId"`
	OfferID   string `json:"offerId"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

type ErasmumuOffer struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Domain    string  `json:"domain"`
	Available bool    `json:"available"`
}

// POST /internship
func registerInternship(w http.ResponseWriter, r *http.Request) {
	var input struct {
		StudentID int    `json:"studentId"`
		OfferID   string `json:"offerId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Get Student
	var student Student
	err := db.QueryRow(context.Background(),
		"SELECT id, firstname, name, domain FROM students WHERE id=$1", input.StudentID).Scan(&student.ID, &student.Firstname, &student.Name, &student.Domain)
	if err != nil {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	// 2. Get Offer from Erasmumu
	// Use environment variable for service URL if available, else default to docker service name or localhost depending on context
	erasmumuURL := "http://erasmumu:8080" 
    if os.Getenv("ERASMUMU_URL") != "" {
        erasmumuURL = os.Getenv("ERASMUMU_URL")
    }

	resp, err := http.Get(fmt.Sprintf("%s/offer/%s", erasmumuURL, input.OfferID))
	if err != nil {
		http.Error(w, "Failed to contact Erasmumu service: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		http.Error(w, "Offer not found", http.StatusNotFound)
		return
	}

	var offer ErasmumuOffer
	if err := json.NewDecoder(resp.Body).Decode(&offer); err != nil {
		http.Error(w, "Failed to parse offer", http.StatusInternalServerError)
		return
	}

	// 3. Validation Logic
	status := "approved"
	message := "Student successfully registered"

	if !offer.Available {
        status = "rejected"
        message = "Offer is not available"
    } else if !strings.EqualFold(student.Domain, offer.Domain) {
		status = "rejected"
		message = "Offer domain doesn't match" 
	}

	// 4. Save Registration
	var regID int
	err = db.QueryRow(context.Background(),
		"INSERT INTO registrations (student_id, offer_id, status, message) VALUES ($1, $2, $3, $4) RETURNING id",
		input.StudentID, input.OfferID, status, message).Scan(&regID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := Registration{
		ID:        regID,
		StudentID: input.StudentID,
		OfferID:   input.OfferID,
		Status:    status,
		Message:   message,
	}
    
	jsonResponse(w, http.StatusCreated, response)
}

// GET /internship/:id
func getRegistration(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var reg Registration
	err = db.QueryRow(context.Background(),
		"SELECT id, student_id, offer_id, status, message FROM registrations WHERE id=$1", id).Scan(&reg.ID, &reg.StudentID, &reg.OfferID, &reg.Status, &reg.Message)
	if err != nil {
		http.Error(w, "Registration not found", http.StatusNotFound)
		return
	}

	jsonResponse(w, http.StatusOK, reg)
}
