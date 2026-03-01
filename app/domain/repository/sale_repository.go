package repository

import (
	"context"
	"pharmacy-pos/api/app/domain/model"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SaleRepository interface {
	Create(ctx context.Context, sale *model.Sale) (*model.Sale, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.Sale, error)
	FindByClientID(ctx context.Context, clientID string, page int, limit int) ([]model.Sale, int64, error)
	FindByPatientID(ctx context.Context, patientID primitive.ObjectID, page int, limit int) ([]model.Sale, int64, error)
	FindByDateRange(ctx context.Context, clientID string, from time.Time, to time.Time) ([]model.Sale, error)
	FindControlledSales(ctx context.Context, clientID string, classification model.DrugClassification, from time.Time, to time.Time) ([]model.Sale, error)
}
