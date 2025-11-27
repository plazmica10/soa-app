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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type tourExecRepo interface {
	CreateExecution(ctx context.Context, exec *model.TourExecution) (*model.TourExecution, error)
	GetActiveExecution(ctx context.Context, touristId string, tourId primitive.ObjectID) (*model.TourExecution, error)
	UpdateExecution(ctx context.Context, exec *model.TourExecution) error
}

func RegisterExecutionRoutes(authRouter *mux.Router, repo tourExecRepo) {
	if authRouter != nil {
		authRouter.HandleFunc("/executions", createExecution(repo)).Methods("POST")
		authRouter.HandleFunc("/executions/{tourId}/active", getActiveExecution(repo)).Methods("GET")
		authRouter.HandleFunc("/executions/{execId}", updateExecution(repo)).Methods("PUT")
	}
}

type createExecutionRequest struct {
	TourID string `json:"tourId"`
}

func createExecution(repo tourExecRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a := auth.GetAuth(r)
		if a == nil || a.UserID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req createExecutionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		tourObjID, err := primitive.ObjectIDFromHex(req.TourID)
		if err != nil {
			http.Error(w, "invalid tourId", http.StatusBadRequest)
			return
		}

		exec := &model.TourExecution{
			TourID:    tourObjID,
			TouristID: a.UserID,
			Status:    model.ExecutionActive,
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		created, err := repo.CreateExecution(ctx, exec)
		if err != nil {
			log.Println("create execution error:", err)
			http.Error(w, "failed to create execution", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
	}
}

func getActiveExecution(repo tourExecRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a := auth.GetAuth(r)
		if a == nil || a.UserID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		tourId := vars["tourId"]
		if tourId == "" {
			http.Error(w, "tourId required", http.StatusBadRequest)
			return
		}

		tourObjID, err := primitive.ObjectIDFromHex(tourId)
		if err != nil {
			http.Error(w, "invalid tourId", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		exec, err := repo.GetActiveExecution(ctx, a.UserID, tourObjID)
		if err != nil {
			http.Error(w, "active execution not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(exec)
	}
}

type updateExecutionRequest struct {
	Status          string   `json:"status"`
	CompletedPoints []string `json:"completedPoints"`
}

func updateExecution(repo tourExecRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a := auth.GetAuth(r)
		if a == nil || a.UserID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		execId := vars["execId"]
		if execId == "" {
			http.Error(w, "execution ID required", http.StatusBadRequest)
			return
		}

		objID, err := primitive.ObjectIDFromHex(execId)
		if err != nil {
			http.Error(w, "invalid execution ID", http.StatusBadRequest)
			return
		}

		var req updateExecutionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		status := model.ExecutionStatus(req.Status)

		// Popunjavanje CompletedPoints
		completedPoints := make([]model.CompletedPoint, 0, len(req.CompletedPoints))
		now := time.Now().UTC()
		for _, kpID := range req.CompletedPoints {
			id, err := primitive.ObjectIDFromHex(kpID)
			if err != nil {
				continue
			}
			completedPoints = append(completedPoints, model.CompletedPoint{
				KeyPointID: id,
				ReachedAt:  now,
			})
		}

		// Ako je završena ili napuštena, postavi FinishedAt
		var finishedAt *time.Time
		if status == model.ExecutionCompleted || status == model.ExecutionAbandoned {
			finishedAt = &now
		}

		exec := &model.TourExecution{
			ID:              objID,
			Status:          status,
			LastActivity:    now,
			CompletedPoints: completedPoints,
			FinishedAt:      finishedAt,
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := repo.UpdateExecution(ctx, exec); err != nil {
			log.Println("update execution error:", err)
			http.Error(w, "failed to update execution", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
