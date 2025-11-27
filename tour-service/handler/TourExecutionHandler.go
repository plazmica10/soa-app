package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"tour-service/auth"
	"tour-service/model"
	"tour-service/utils"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type tourExecRepo interface {
	CreateExecution(ctx context.Context, exec *model.TourExecution) (*model.TourExecution, error)
	GetActiveExecution(ctx context.Context, touristId string, tourId primitive.ObjectID) (*model.TourExecution, error)
	UpdateExecution(ctx context.Context, exec *model.TourExecution) error
	AddLocation(ctx context.Context, execId primitive.ObjectID, loc model.Location) error
	CompletePoint(ctx context.Context, execId primitive.ObjectID, cp model.CompletedPoint) error
	GetKeyPointsByTour(ctx context.Context, tourId primitive.ObjectID) ([]model.KeyPoint, error)
	GetTourByID(ctx context.Context, tourId string) (*model.Tour, error)
	GetExecutionByID(ctx context.Context, execId primitive.ObjectID) (*model.TourExecution, error)
}

func RegisterExecutionRoutes(authRouter *mux.Router, execRepo tourExecRepo) {
	if authRouter != nil {
		authRouter.HandleFunc("/executions", createExecution(execRepo)).Methods("POST")
		authRouter.HandleFunc("/executions/{tourId}/active", getActiveExecution(execRepo)).Methods("GET")
		authRouter.HandleFunc("/executions/{execId}", updateExecution(execRepo)).Methods("PUT")
		authRouter.HandleFunc("/executions/{execId}/location", addLocation(execRepo)).Methods("POST")
		authRouter.HandleFunc("/executions/{execId}/complete", completePoint(execRepo)).Methods("POST")
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

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		tour, err := repo.GetTourByID(r.Context(), req.TourID)
		if err != nil {
			http.Error(w, "tour not found", http.StatusNotFound)
			return
		}

		if tour.Status == "draft" {
			http.Error(w, "tour is not in draft status", http.StatusBadRequest)
			return
		}

		exec := &model.TourExecution{
			TourID:    tourObjID,
			TouristID: a.UserID,
			Status:    model.ExecutionActive,
		}

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

type addLocationRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func addLocation(repo tourExecRepo) http.HandlerFunc {
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

		var req addLocationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		exec, err := repo.GetExecutionByID(ctx, objID)
		if err != nil {
			http.Error(w, "execution not found", http.StatusNotFound)
			return
		}

		kps, err := repo.GetKeyPointsByTour(ctx, exec.TourID)
		if err != nil {
			http.Error(w, "failed to get keypoints", http.StatusInternalServerError)
			return
		}

		nearAny := false
		for _, kp := range kps {
			if utils.IsNearby(req.Latitude, req.Longitude, kp.Latitude, kp.Longitude) {
				nearAny = true
				break
			}
		}
		if !nearAny {
			http.Error(w, "location too far from any keypoint", http.StatusBadRequest)
			return
		}

		loc := model.Location{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Timestamp: time.Now().UTC(),
		}

		if err := repo.AddLocation(ctx, objID, loc); err != nil {
			log.Println("add location error:", err)
			http.Error(w, "failed to add location", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

type completePointRequest struct {
	KeyPointID string `json:"keyPointId"`
}

func completePoint(repo tourExecRepo) http.HandlerFunc {
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

		var req completePointRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		kpID, err := primitive.ObjectIDFromHex(req.KeyPointID)
		if err != nil {
			http.Error(w, "invalid keyPointId", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		exec, err := repo.GetActiveExecution(ctx, a.UserID, objID)
		if err != nil {
			http.Error(w, "active execution not found", http.StatusNotFound)
			return
		}

		kps, err := repo.GetKeyPointsByTour(ctx, exec.TourID)
		if err != nil {
			http.Error(w, "failed to get keypoints", http.StatusInternalServerError)
			return
		}

		var keypoint model.KeyPoint
		found := false
		for _, kp := range kps {
			if kp.ID == kpID {
				keypoint = kp
				found = true
				break
			}
		}
		if !found {
			http.Error(w, "keypoint not found", http.StatusBadRequest)
			return
		}

		if len(exec.Locations) == 0 {
			http.Error(w, "no location recorded", http.StatusBadRequest)
			return
		}
		lastLoc := exec.Locations[len(exec.Locations)-1]

		if !utils.IsNearby(lastLoc.Latitude, lastLoc.Longitude, keypoint.Latitude, keypoint.Longitude) {
			http.Error(w, "too far from keypoint", http.StatusBadRequest)
			return
		}

		cp := model.CompletedPoint{
			KeyPointID: kpID,
			ReachedAt:  time.Now().UTC(),
		}

		if err := repo.CompletePoint(ctx, objID, cp); err != nil {
			log.Println("complete point error:", err)
			http.Error(w, "failed to complete point", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
