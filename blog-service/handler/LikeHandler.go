package handler

import (
    "encoding/json"
    "net/http"
    "strings"

    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson/primitive"

    "blog-service/model"
    "blog-service/repository"
)

type likeHandler struct {
    repo     *repository.LikeRepository
    blogRepo *repository.BlogRepository
}

// RegisterLikeRoutes registers like endpoints
func RegisterLikeRoutes(r *mux.Router, lr *repository.LikeRepository, br *repository.BlogRepository) {
    h := &likeHandler{repo: lr, blogRepo: br}
    r.HandleFunc("/blogs/{id}/likes", h.createLike).Methods("POST")
    r.HandleFunc("/blogs/{id}/likes", h.deleteLike).Methods("DELETE")
}

type likeReq struct {
    UserID string `json:"user_id"`
}

func (h *likeHandler) createLike(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    if id == "" {
        http.Error(w, "blog id required", http.StatusBadRequest)
        return
    }
    var in likeReq
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    if in.UserID == "" {
        http.Error(w, "user_id required", http.StatusBadRequest)
        return
    }
    blogOID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        http.Error(w, "invalid blog id", http.StatusBadRequest)
        return
    }
    userOID, err := primitive.ObjectIDFromHex(in.UserID)
    if err != nil {
        http.Error(w, "invalid user id", http.StatusBadRequest)
        return
    }
    // ensure blog exists
    if _, err := h.blogRepo.GetByID(r.Context(), id); err != nil {
        http.Error(w, "blog not found", http.StatusNotFound)
        return
    }
    l := model.Like{BlogID: blogOID, UserID: userOID}
    err = h.repo.Create(r.Context(), &l)
    if err != nil {
        // handle duplicate (already liked) as idempotent
        if mongoErrIsDuplicate(err) {
            w.WriteHeader(http.StatusOK)
            w.Write([]byte("already liked"))
            return
        }
        http.Error(w, "failed to create like", http.StatusInternalServerError)
        return
    }
    // increment blog likes_count
    _ = h.blogRepo.UpdateLikesCount(r.Context(), blogOID, 1)
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(l)
}

func (h *likeHandler) deleteLike(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    if id == "" {
        http.Error(w, "blog id required", http.StatusBadRequest)
        return
    }
    var in likeReq
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    if in.UserID == "" {
        http.Error(w, "user_id required", http.StatusBadRequest)
        return
    }
    blogOID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        http.Error(w, "invalid blog id", http.StatusBadRequest)
        return
    }
    userOID, err := primitive.ObjectIDFromHex(in.UserID)
    if err != nil {
        http.Error(w, "invalid user id", http.StatusBadRequest)
        return
    }
    // delete like
    deleted, err := h.repo.DeleteByBlogAndUser(r.Context(), blogOID, userOID)
    if err != nil {
        http.Error(w, "failed to delete like", http.StatusInternalServerError)
        return
    }
    if deleted > 0 {
        _ = h.blogRepo.UpdateLikesCount(r.Context(), blogOID, -1)
		http.Error(w, "like removed", http.StatusOK)
		return
    }
    w.WriteHeader(http.StatusNoContent)
}

// mongoErrIsDuplicate checks if err is a duplicate key error (11000)
func mongoErrIsDuplicate(err error) bool {
    // avoid importing mongo types here; do a string check as lightweight approach
    if err == nil {
        return false
    }
    return strings.Contains(err.Error(), "E11000") || strings.Contains(err.Error(), "duplicate key")
}
