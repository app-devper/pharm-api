package repo

import (
	"context"
	"pharmacy-pos/api/app/domain/model"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type productRepo struct {
	col *mongo.Collection
}

func NewProductRepo(db *mongo.Database) *productRepo {
	return &productRepo{col: db.Collection("products")}
}

func (r *productRepo) Create(ctx context.Context, product *model.Product) (*model.Product, error) {
	result, err := r.col.InsertOne(ctx, product)
	if err != nil {
		return nil, err
	}
	product.ID = result.InsertedID.(primitive.ObjectID)
	return product, nil
}

func (r *productRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Product, error) {
	var product model.Product
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&product)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepo) FindByClientID(ctx context.Context, clientID string, search string, page int, limit int) ([]model.Product, int64, error) {
	filter := bson.M{"clientId": clientID, "status": "ACTIVE"}
	if search != "" {
		escaped := regexp.QuoteMeta(search)
		filter["$or"] = []bson.M{
			{"tradeName": bson.M{"$regex": escaped, "$options": "i"}},
			{"genericName": bson.M{"$regex": escaped, "$options": "i"}},
			{"barcode": bson.M{"$regex": escaped, "$options": "i"}},
		}
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().SetSkip(skip).SetLimit(int64(limit)).SetSort(bson.D{{Key: "tradeName", Value: 1}})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var products []model.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *productRepo) FindByBarcode(ctx context.Context, clientID string, barcode string) (*model.Product, error) {
	var product model.Product
	err := r.col.FindOne(ctx, bson.M{"clientId": clientID, "barcode": barcode}).Decode(&product)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepo) Update(ctx context.Context, product *model.Product) error {
	product.UpdatedDate = time.Now()
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": product.ID}, bson.M{"$set": product})
	return err
}

func (r *productRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"status": "INACTIVE", "updatedDate": time.Now()}})
	return err
}
