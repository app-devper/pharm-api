package repository

import (
	"context"
	"pharmacy-pos/api/app/domain/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PatientRepository interface {
	Create(ctx context.Context, patient *model.Patient) (*model.Patient, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.Patient, error)
	FindByIDCard(ctx context.Context, idCard string) (*model.Patient, error)
	FindAll(ctx context.Context, search string, page int, limit int) ([]model.Patient, int64, error)
	Update(ctx context.Context, patient *model.Patient) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}
