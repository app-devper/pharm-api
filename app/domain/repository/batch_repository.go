package repository

import (
	"context"
	"pharmacy-pos/api/app/domain/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BatchRepository interface {
	Create(ctx context.Context, batch *model.Batch) (*model.Batch, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.Batch, error)
	FindByProductID(ctx context.Context, productID primitive.ObjectID) ([]model.Batch, error)
	FindByProductIDFEFO(ctx context.Context, productID primitive.ObjectID) ([]model.Batch, error)
	FindExpiringBatches(ctx context.Context, clientID string, daysAhead int) ([]model.Batch, error)
	FindLowStock(ctx context.Context, clientID string) ([]model.Batch, error)
	UpdateQuantity(ctx context.Context, id primitive.ObjectID, quantity int) error
	Update(ctx context.Context, id primitive.ObjectID, update map[string]interface{}) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}
