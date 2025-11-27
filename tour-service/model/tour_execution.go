package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ExecutionStatus string

const (
	ExecutionActive    ExecutionStatus = "active"
	ExecutionCompleted ExecutionStatus = "completed"
	ExecutionAbandoned ExecutionStatus = "abandoned"
)

type TourExecution struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TourID          primitive.ObjectID `bson:"tourId" json:"tourId"`
	TouristID       string             `bson:"touristId" json:"touristId"`
	StartedAt       time.Time          `bson:"startedAt" json:"startedAt"`
	FinishedAt      *time.Time         `bson:"finishedAt,omitempty" json:"finishedAt,omitempty"`
	Status          ExecutionStatus    `bson:"status" json:"status"`
	LastActivity    time.Time          `bson:"lastActivity" json:"lastActivity"`
	CompletedPoints []CompletedPoint   `bson:"completedPoints" json:"completedPoints"`
}

type CompletedPoint struct {
	KeyPointID primitive.ObjectID `bson:"keyPointId" json:"keyPointId"`
	ReachedAt  time.Time          `bson:"reachedAt" json:"reachedAt"`
}
