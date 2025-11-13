package handler

import (
	"encoding/json"
	"net/http"
	"time"
	"log"

	"github.com/gorilla/mux"

	"blog-service/model"
	"blog-service/repository"
)

type blogHandler struct {
	repo *repository.BlogRepository
}

// RegisterRoutes registers blog routes on the given router
func RegisterRoutes(r *mux.Router, repo *repository.BlogRepository) {
	h := &blogHandler{repo: repo}
	r.HandleFunc("/blogs", h.createBlog).Methods("POST")
	r.HandleFunc("/blogs", h.listBlogs).Methods("GET")
	r.HandleFunc("/blogs/{id}", h.getBlog).Methods("GET")
}

func (h *blogHandler) createBlog(w http.ResponseWriter, r *http.Request) {
	var in model.Blog
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	// basic validation
	if in.Title == "" || in.Description == "" {
		http.Error(w, "title and description are required", http.StatusBadRequest)
		return
	}
	in.CreatedAt = time.Now().UTC()
    if err := h.repo.Create(r.Context(), &in); err != nil {
        log.Printf("create blog error: %v", err)
        http.Error(w, "failed to create blog: "+err.Error(), http.StatusInternalServerError)
        return
    }
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(in)
}

func (h *blogHandler) listBlogs(w http.ResponseWriter, r *http.Request) {
	blogs, err := h.repo.GetAll(r.Context())
	if err != nil {
		http.Error(w, "failed to list blogs", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blogs)
}

func (h *blogHandler) getBlog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	b, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}
