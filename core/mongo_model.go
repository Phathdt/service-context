package core

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MongoModel struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt,omitempty"`
}
