package core

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MongoModel struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"created_at" bson:"createdAt,omitempty"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updatedAt,omitempty"`
}
