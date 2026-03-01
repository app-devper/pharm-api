package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Setting struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ClientID    string             `json:"clientId" bson:"clientId"`
	Key         string             `json:"key" bson:"key"`
	Value       string             `json:"value" bson:"value"`
	UpdatedBy   string             `json:"updatedBy" bson:"updatedBy"`
	UpdatedDate time.Time          `json:"updatedDate" bson:"updatedDate"`
}
