package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username     string             `bson:"username" json:"username"`
	Password     string             `bson:"password" json:"password"`
	Email        string             `bson:"email" json:"email"`
	Name         string             `bson:"name" json:"name"`
	Surname      string             `bson:"surname" json:"surname"`
	Roles        []string           `bson:"roles" json:"roles"`
	Address      Address            `bson:"address" json:"address"`
	ProfileImage string             `bson:"profile_image,omitempty" json:"profile_image,omitempty"`
	Biography    string             `bson:"biography,omitempty" json:"biography,omitempty"`
	Motto        string             `bson:"motto,omitempty" json:"motto,omitempty"`
	IsBlocked    bool               `bson:"is_blocked" json:"is_blocked"`
}
