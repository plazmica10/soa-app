package model

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

// Comment represents a comment left on a blog post.
type Comment struct {
    ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    BlogID       primitive.ObjectID `bson:"blog_id" json:"blog_id"`
    AuthorID     primitive.ObjectID `bson:"author_id,omitempty" json:"author_id,omitempty"`
    AuthorName   string             `bson:"author_name" json:"author_name"`
    Text         string             `bson:"text" json:"text"`
    CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
    LastEditedAt *time.Time         `bson:"last_edited_at,omitempty" json:"last_edited_at,omitempty"`
}
