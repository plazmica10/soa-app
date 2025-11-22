package handler

import (
	"encoding/json"
	"net/http"
	"time"
	"log"
	"os"
	"fmt"

	"github.com/gorilla/mux"

	"blog-service/auth"
	"blog-service/model"
	"blog-service/repository"
)

type blogHandler struct {
	repo *repository.BlogRepository
}

// RegisterRoutes registers blog routes. Public routes go on 'public',
// protected routes (requiring auth) go on 'authRouter'.
func RegisterRoutes(public *mux.Router, authRouter *mux.Router, repo *repository.BlogRepository) {
	h := &blogHandler{repo: repo}
	// protected (reads require authentication/follow checks)
	if authRouter != nil {
		authRouter.HandleFunc("/blogs", h.listBlogs).Methods("GET")
		authRouter.HandleFunc("/blogs/{id}", h.getBlog).Methods("GET")
	} else {
		// fallback to public if no auth router provided (handlers will still enforce auth)
		public.HandleFunc("/blogs", h.listBlogs).Methods("GET")
		public.HandleFunc("/blogs/{id}", h.getBlog).Methods("GET")
	}
	// protected
	if authRouter != nil {
		authRouter.HandleFunc("/blogs", h.createBlog).Methods("POST")
	}
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
	// attach author info from JWT if available
	if a := auth.GetAuth(r); a != nil {
		// try to set AuthorID and AuthorName when posting a blog
		if a.UserID != "" {
			in.AuthorID = a.UserID
		}
		if in.AuthorName == "" {
			in.AuthorName = a.Username
		}
	}
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
	// Only authenticated users may read blogs of users they follow.
	a := auth.GetAuth(r)
	if a == nil || a.UserID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get list of users the current user follows from follower-service
	following, err := h.getFollowingIDs(r, a.UserID)
	if err != nil {
		http.Error(w, "failed to get following: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// include self so users see their own blogs
	following = append(following, a.UserID)

	blogs, err := h.repo.GetByAuthorIDs(r.Context(), following)
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
	// Enforce that only users who follow the author (or the author themself) can read the blog
	a := auth.GetAuth(r)
	if a == nil || a.UserID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if b.AuthorID != a.UserID {
		// check following list
		following, err := h.getFollowingIDs(r, a.UserID)
		if err != nil {
			http.Error(w, "failed to verify following: "+err.Error(), http.StatusInternalServerError)
			return
		}
		allowed := false
		for _, id := range following {
			if id == b.AuthorID {
				allowed = true
				break
			}
		}
		if !allowed {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}

// getFollowingIDs calls follower-service to retrieve list of user IDs the given user is following.
func (h *blogHandler) getFollowingIDs(r *http.Request, userID string) ([]string, error) {
	base := os.Getenv("FOLLOWER_SERVICE_URL")
	if base == "" {
		base = "http://follower-service:8082"
	}
	url := fmt.Sprintf("%s/following/%s", base, userID)
	req, err := http.NewRequestWithContext(r.Context(), "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("follower service returned %d", resp.StatusCode)
	}
	var out []string
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}
