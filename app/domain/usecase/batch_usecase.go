package usecase

import (
	"context"
	"pharmacy-pos/api/app/domain/model"
	"pharmacy-pos/api/app/domain/repository"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BatchUsecase struct {
	batchRepo   repository.BatchRepository
	productRepo repository.ProductRepository
}

func NewBatchUsecase(batchRepo repository.BatchRepository, productRepo repository.ProductRepository) *BatchUsecase {
	return &BatchUsecase{batchRepo: batchRepo, productRepo: productRepo}
}

func (u *BatchUsecase) ReceiveGoods(ctx context.Context, batch *model.Batch, userID string) (*model.Batch, error) {
	batch.CreatedBy = userID
	batch.CreatedDate = time.Now()
	batch.UpdatedBy = userID
	batch.UpdatedDate = time.Now()
	batch.ReceivedAt = time.Now()
	return u.batchRepo.Create(ctx, batch)
}

func (u *BatchUsecase) GetByID(ctx context.Context, id string) (*model.Batch, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return u.batchRepo.FindByID(ctx, oid)
}

func (u *BatchUsecase) GetByProductID(ctx context.Context, productID string) ([]model.Batch, error) {
	oid, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return nil, err
	}
	return u.batchRepo.FindByProductID(ctx, oid)
}

func (u *BatchUsecase) GetByProductIDFEFO(ctx context.Context, productID string) ([]model.Batch, error) {
	oid, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return nil, err
	}
	return u.batchRepo.FindByProductIDFEFO(ctx, oid)
}

func (u *BatchUsecase) GetExpiringBatches(ctx context.Context, clientID string, daysAhead int) ([]model.Batch, error) {
	if daysAhead <= 0 {
		daysAhead = 180
	}
	return u.batchRepo.FindExpiringBatches(ctx, clientID, daysAhead)
}

func (u *BatchUsecase) GetLowStock(ctx context.Context, clientID string) ([]model.Batch, error) {
	return u.batchRepo.FindLowStock(ctx, clientID)
}

func (u *BatchUsecase) UpdateQuantity(ctx context.Context, id string, quantity int) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	return u.batchRepo.UpdateQuantity(ctx, oid, quantity)
}

func (u *BatchUsecase) Update(ctx context.Context, id string, update map[string]interface{}) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	return u.batchRepo.Update(ctx, oid, update)
}

func (u *BatchUsecase) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	return u.batchRepo.Delete(ctx, oid)
}
