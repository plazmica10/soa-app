package model

import "time"

// Blog represents a blog post. Description is stored as raw markdown.
import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Blog struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    Images      []string  `json:"images,omitempty"`
}
