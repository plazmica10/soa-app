package handler

import (
	"encoding/json"
	"net/http"

	// "stakeholders-service/model"
	"stakeholders-service/repository"

	"github.com/gorilla/mux"
)

func RegisterRoutes(router *mux.Router, userRepo *repository.UserRepository) {
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	router.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		users, err := userRepo.GetAll(ctx)
		if err != nil {
			http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(users)
	}).Methods("GET")
}
