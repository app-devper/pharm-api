package usecase

import (
	"context"
	"pharmacy-pos/api/app/domain/model"
	"pharmacy-pos/api/app/domain/repository"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PatientUsecase struct {
	repo repository.PatientRepository
}

func NewPatientUsecase(repo repository.PatientRepository) *PatientUsecase {
	return &PatientUsecase{repo: repo}
}

func (u *PatientUsecase) Create(ctx context.Context, patient *model.Patient, userID string) (*model.Patient, error) {
	patient.CreatedBy = userID
	patient.CreatedDate = time.Now()
	patient.UpdatedBy = userID
	patient.UpdatedDate = time.Now()
	return u.repo.Create(ctx, patient)
}

func (u *PatientUsecase) GetByID(ctx context.Context, id string) (*model.Patient, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return u.repo.FindByID(ctx, oid)
}

func (u *PatientUsecase) GetByClientID(ctx context.Context, clientID string, search string, page int, limit int) ([]model.Patient, int64, error) {
	return u.repo.FindByClientID(ctx, clientID, search, page, limit)
}

func (u *PatientUsecase) Update(ctx context.Context, patient *model.Patient, userID string) error {
	patient.UpdatedBy = userID
	patient.UpdatedDate = time.Now()
	return u.repo.Update(ctx, patient)
}

func (u *PatientUsecase) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	return u.repo.Delete(ctx, oid)
}
