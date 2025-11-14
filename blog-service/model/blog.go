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
    // Author information (optional when anonymous or set from JWT)
    AuthorID   string    `bson:"author_id,omitempty" json:"author_id,omitempty"`
    AuthorName string    `bson:"author_name,omitempty" json:"author_name,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
    Images      []string  `json:"images,omitempty"`
    LikesCount  int       `bson:"likes_count,omitempty" json:"likes_count,omitempty"`
}
