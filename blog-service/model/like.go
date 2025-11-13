package model

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Like struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    BlogID    primitive.ObjectID `bson:"blog_id" json:"blog_id"`
    UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
    CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
