package usecase

import (
	"context"
	"pharmacy-pos/api/app/domain/model"
	"pharmacy-pos/api/app/domain/repository"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductUsecase struct {
	repo repository.ProductRepository
}

func NewProductUsecase(repo repository.ProductRepository) *ProductUsecase {
	return &ProductUsecase{repo: repo}
}

func (u *ProductUsecase) Create(ctx context.Context, product *model.Product, userID string) (*model.Product, error) {
	product.CreatedBy = userID
	product.CreatedDate = time.Now()
	product.UpdatedBy = userID
	product.UpdatedDate = time.Now()
	product.Status = "ACTIVE"
	return u.repo.Create(ctx, product)
}

func (u *ProductUsecase) GetByID(ctx context.Context, id string) (*model.Product, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return u.repo.FindByID(ctx, oid)
}

func (u *ProductUsecase) GetByClientID(ctx context.Context, clientID string, search string, page int, limit int) ([]model.Product, int64, error) {
	return u.repo.FindByClientID(ctx, clientID, search, page, limit)
}

func (u *ProductUsecase) GetByBarcode(ctx context.Context, clientID string, barcode string) (*model.Product, error) {
	return u.repo.FindByBarcode(ctx, clientID, barcode)
}

func (u *ProductUsecase) Update(ctx context.Context, product *model.Product, userID string) error {
	product.UpdatedBy = userID
	product.UpdatedDate = time.Now()
	return u.repo.Update(ctx, product)
}

func (u *ProductUsecase) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	return u.repo.Delete(ctx, oid)
}
