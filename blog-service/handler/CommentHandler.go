package handler

import (
    "encoding/json"
    "net/http"

    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson/primitive"

    "blog-service/auth"
    "blog-service/model"
    "blog-service/repository"
)

type commentHandler struct {
    repo     *repository.CommentRepository
    blogRepo *repository.BlogRepository
}

// RegisterCommentRoutes registers comment endpoints
func RegisterCommentRoutes(public *mux.Router, authRouter *mux.Router, cr *repository.CommentRepository, br *repository.BlogRepository) {
    h := &commentHandler{repo: cr, blogRepo: br}
    // public
    public.HandleFunc("/blogs/{id}/comments", h.listComments).Methods("GET")
    // protected
    if authRouter != nil {
        authRouter.HandleFunc("/blogs/{id}/comments", h.createComment).Methods("POST")
        authRouter.HandleFunc("/blogs/{id}/comments/{cid}", h.updateComment).Methods("PATCH")
    }
}

type createCommentReq struct {
    // AuthorID may be omitted when the request is authenticated; we'll use JWT user id.
    AuthorID   string `json:"author_id,omitempty"`
    AuthorName string `json:"author_name"`
    Text       string `json:"text"`
}

func (h *commentHandler) createComment(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    if id == "" {
        http.Error(w, "blog id required", http.StatusBadRequest)
        return
    }
    // validate blog exists
    if _, err := h.blogRepo.GetByID(r.Context(), id); err != nil {
        http.Error(w, "blog not found", http.StatusNotFound)
        return
    }
    var in createCommentReq
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    if in.Text == "" {
        http.Error(w, "text is required", http.StatusBadRequest)
        return
    }
    // parse ids
    blogOID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        http.Error(w, "invalid blog id", http.StatusBadRequest)
        return
    }
    var authorOID primitive.ObjectID
    // prefer authenticated user id over provided author_id
    if a := auth.GetAuth(r); a != nil && a.UserID != "" {
        if au, err := primitive.ObjectIDFromHex(a.UserID); err == nil {
            authorOID = au
        }
        // use username from token if author_name not provided
        if in.AuthorName == "" {
            in.AuthorName = a.Username
        }
    } else {
        if in.AuthorID != "" {
            a, err := primitive.ObjectIDFromHex(in.AuthorID)
            if err == nil {
                authorOID = a
            }
        }
    }
    c := model.Comment{
        BlogID:     blogOID,
        AuthorID:   authorOID,
        AuthorName: in.AuthorName,
        Text:       in.Text,
    }
    if err := h.repo.Create(r.Context(), &c); err != nil {
        http.Error(w, "failed to create comment", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(c)
}

func (h *commentHandler) listComments(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    if id == "" {
        http.Error(w, "blog id required", http.StatusBadRequest)
        return
    }
    blogOID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        http.Error(w, "invalid blog id", http.StatusBadRequest)
        return
    }
    comments, err := h.repo.GetByBlogID(r.Context(), blogOID)
    if err != nil {
        http.Error(w, "failed to list comments", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(comments)
}

type editCommentReq struct {
    Text string `json:"text"`
}

func (h *commentHandler) updateComment(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    blogID := vars["id"]
    cid := vars["cid"]
    if blogID == "" || cid == "" {
        http.Error(w, "blog id and comment id required", http.StatusBadRequest)
        return
    }
    // ensure blog exists
    if _, err := h.blogRepo.GetByID(r.Context(), blogID); err != nil {
        http.Error(w, "blog not found", http.StatusNotFound)
        return
    }
    commentOID, err := primitive.ObjectIDFromHex(cid)
    if err != nil {
        http.Error(w, "invalid comment id", http.StatusBadRequest)
        return
    }
    var in editCommentReq
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    if in.Text == "" {
        http.Error(w, "text is required", http.StatusBadRequest)
        return
    }
    updated, err := h.repo.UpdateText(r.Context(), commentOID, in.Text)
    if err != nil {
        http.Error(w, "failed to update comment", http.StatusInternalServerError)
        return
    }
    // ensure comment belongs to blog
    if updated.BlogID.Hex() != blogID {
        http.Error(w, "comment does not belong to blog", http.StatusBadRequest)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(updated)
}
