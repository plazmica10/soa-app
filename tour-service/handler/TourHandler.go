package handler

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "time"

    "github.com/gorilla/mux"
    "tour-service/model"
)

type tourRepo interface {
    CreateTour(ctx context.Context, t *model.Tour) (*model.Tour, error)
    GetToursByAuthor(ctx context.Context, authorId string) ([]model.Tour, error)
}

func RegisterRoutes(r *mux.Router, repo tourRepo) {
    r.HandleFunc("/tours", createTour(repo)).Methods("POST")
    r.HandleFunc("/tours/author/{authorId}", listToursByAuthor(repo)).Methods("GET")
}

type createTourRequest struct {
    AuthorID    string   `json:"authorId"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Difficulty  string   `json:"difficulty"`
    Tags        []string `json:"tags"`
}

//TODO: use authentication to get authorId, need authentication service
func createTour(repo tourRepo) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req createTourRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "invalid request", http.StatusBadRequest)
            return
        }
        if req.AuthorID == "" || req.Name == "" {
            http.Error(w, "authorId and name are required", http.StatusBadRequest)
            return
        }
        t := &model.Tour{
            AuthorID:    req.AuthorID,
            Name:        req.Name,
            Description: req.Description,
            Difficulty:  req.Difficulty,
            Tags:        req.Tags,
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
