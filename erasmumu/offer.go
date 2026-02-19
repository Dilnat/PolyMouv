package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Offer struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title" json:"title"`
	Link      string             `bson:"link" json:"link"`
	City      string             `bson:"city" json:"city"`
	Domain    string             `bson:"domain" json:"domain"`
	Salary    float64            `bson:"salary" json:"salary"`
	StartDate string             `bson:"startDate" json:"startDate"` // Using string for simplicity as per common lab practices, or time.Time? Let's use string based on prompt "StartDate" usually implies date.
	EndDate   string             `bson:"endDate" json:"endDate"`
	Available bool               `bson:"available" json:"available"`
}

// Helper for JSON responses
func jsonResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// POST /offer
func createOffer(w http.ResponseWriter, r *http.Request) {
	var o Offer
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

    // New offers are typically available by default if not specified, 
    // but we respect the payload.
	o.ID = primitive.NewObjectID()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, o)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusCreated, o)
}

// GET /offer/{id}
func getOffer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var o Offer
    // Rule: An offer must not be returned if available is false
	err = collection.FindOne(ctx, bson.M{"_id": id, "available": true}).Decode(&o)
	if err != nil {
		http.Error(w, "Offer not found", http.StatusNotFound) // Could distinguish err type, but keep it simple
		return
	}

	jsonResponse(w, http.StatusOK, o)
}

// GET /offer?domain=<domain>&city=<city>
func getOffers(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	city := r.URL.Query().Get("city")

	filter := bson.M{"available": true} // Rule: An offer must not be returned if available is false
	if domain != "" {
		filter["domain"] = domain
	}
	if city != "" {
		filter["city"] = city
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var offers []Offer
	if err = cursor.All(ctx, &offers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
    
    // Return empty array instead of null
    if offers == nil {
        offers = []Offer{}
    }

	jsonResponse(w, http.StatusOK, offers)
}

// PUT /offer/{id}
func updateOffer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var o Offer
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

    // Note: If fields are zero-value in `o`, they will overwrite. 
    // Usually we use a separate struct or map for updates, but for this lab, replacing document logic is acceptable.
    // However, primitive.ObjectID in `o` is zero if not passed, we should exclude it from update or handle it.
    // simpler to just $set specific fields or use 'bson:",omitempty"' in struct which we did for ID.
    // But struct values for update might need cleaner handling.
    // Let's assume the user sends the full object or we accept wiping missing fields (PUT semantics).
    
    // Important: We need to ensure we don't accidentally Unset _id or change it.
    // The struct tag `bson:"_id,omitempty"` might cause issues if we pass the whole struct to $set.
    // Let's trust the decoder/driver behavior for now or be explicit.
    // Being explicit is safer.
    
    updateData := bson.M{
        "title": o.Title,
        "link": o.Link,
        "city": o.City,
        "domain": o.Domain,
        "salary": o.Salary,
        "startDate": o.StartDate,
        "endDate": o.EndDate,
        "available": o.Available,
    }

	result, err := collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": updateData})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
    
    if result.MatchedCount == 0 {
        http.Error(w, "Offer not found", http.StatusNotFound)
        return
    }

    o.ID = id
	jsonResponse(w, http.StatusOK, o)
}

// DELETE /offer/{id}
func deleteOffer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
    
    if result.DeletedCount == 0 {
        http.Error(w, "Offer not found", http.StatusNotFound)
        return
    }

	w.WriteHeader(http.StatusNoContent)
}
