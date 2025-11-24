package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"tour-service/auth"
	"tour-service/model"

	"github.com/gorilla/mux"
)

type tourRepo interface {
	CreateTour(ctx context.Context, t *model.Tour) (*model.Tour, error)
	GetTourByID(ctx context.Context, tourId string) (*model.Tour, error)
	GetToursByAuthor(ctx context.Context, authorId string) ([]model.Tour, error)
	UpdateTour(ctx context.Context, tourId string, authorId string, updates map[string]interface{}) (*model.Tour, error)
}

func RegisterRoutes(public *mux.Router, authRouter *mux.Router, repo tourRepo) {
	// protected routes
	if authRouter != nil {
		authRouter.HandleFunc("/tours", createTour(repo)).Methods("POST")
		authRouter.HandleFunc("/tours/{id}", updateTour(repo)).Methods("PUT")
	}
	// public routes
	public.HandleFunc("/tours/{id}", getTourByID(repo)).Methods("GET")
	public.HandleFunc("/tours/author/{authorId}", listToursByAuthor(repo)).Methods("GET")
}

type createTourRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Difficulty  string   `json:"difficulty"`
	Tags        []string `json:"tags"`
	Status      string   `json:"status"`
}

func createTour(repo tourRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract authenticated user ID from JWT
		a := auth.GetAuth(r)
		if a == nil || a.UserID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req createTourRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		t := &model.Tour{
			AuthorID:    a.UserID,
			Name:        req.Name,
			Description: req.Description,
			Difficulty:  req.Difficulty,
			Tags:        req.Tags,
			Status:      req.Status,
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		created, err := repo.CreateTour(ctx, t)
		if err != nil {
			log.Println("create tour error:", err)
			http.Error(w, "failed to create tour", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
	}
}

type updateTourRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Difficulty  string   `json:"difficulty"`
	Tags        []string `json:"tags"`
	Status      string   `json:"status"`
	Price       float64  `json:"price"`
}

func updateTour(repo tourRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract authenticated user ID from JWT
		a := auth.GetAuth(r)
		if a == nil || a.UserID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		tourId := vars["id"]
		if tourId == "" {
			http.Error(w, "tour id required", http.StatusBadRequest)
			return
		}

		var req updateTourRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		// Build update map
		updates := make(map[string]interface{})
		if req.Name != "" {
			updates["name"] = req.Name
		}
		if req.Description != "" {
			updates["description"] = req.Description
		}
		if req.Difficulty != "" {
			updates["difficulty"] = req.Difficulty
		}
		if req.Tags != nil {
			updates["tags"] = req.Tags
		}
		if req.Status != "" {
			updates["status"] = req.Status
		}
		if req.Price >= 0 {
			updates["price"] = req.Price
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		updated, err := repo.UpdateTour(ctx, tourId, a.UserID, updates)
		if err != nil {
			log.Println("update tour error:", err)
			http.Error(w, "failed to update tour", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updated)
	}
}

func getTourByID(repo tourRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tourId := vars["id"]
		if tourId == "" {
			http.Error(w, "tour id required", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		tour, err := repo.GetTourByID(ctx, tourId)
		if err != nil {
			log.Println("get tour error:", err)
			http.Error(w, "tour not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tour)
	}
}

func listToursByAuthor(repo tourRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		authorId := vars["authorId"]
		if authorId == "" {
			http.Error(w, "authorId required", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		tours, err := repo.GetToursByAuthor(ctx, authorId)
		if err != nil {
			log.Println("list tours error:", err)
			http.Error(w, "failed to list tours", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tours)
	}
}
