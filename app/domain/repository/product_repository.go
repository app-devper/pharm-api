package repository

import (
	"context"
	"pharmacy-pos/api/app/domain/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) (*model.Product, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.Product, error)
	FindByClientID(ctx context.Context, clientID string, search string, page int, limit int) ([]model.Product, int64, error)
	FindByBarcode(ctx context.Context, clientID string, barcode string) (*model.Product, error)
	Update(ctx context.Context, product *model.Product) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}
