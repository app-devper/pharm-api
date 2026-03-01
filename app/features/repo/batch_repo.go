package repo

import (
	"context"
	"pharmacy-pos/api/app/domain/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type batchRepo struct {
	col *mongo.Collection
}

func NewBatchRepo(db *mongo.Database) *batchRepo {
	return &batchRepo{col: db.Collection("batches")}
}

func (r *batchRepo) Create(ctx context.Context, batch *model.Batch) (*model.Batch, error) {
	result, err := r.col.InsertOne(ctx, batch)
	if err != nil {
		return nil, err
	}
	batch.ID = result.InsertedID.(primitive.ObjectID)
	return batch, nil
}

func (r *batchRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Batch, error) {
	var batch model.Batch
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&batch)
	if err != nil {
		return nil, err
	}
	return &batch, nil
}

func (r *batchRepo) FindByProductID(ctx context.Context, productID primitive.ObjectID) ([]model.Batch, error) {
	opts := options.Find().SetSort(bson.D{{Key: "expiryDate", Value: 1}})
	cursor, err := r.col.Find(ctx, bson.M{"productId": productID, "quantity": bson.M{"$gt": 0}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var batches []model.Batch
	if err := cursor.All(ctx, &batches); err != nil {
		return nil, err
	}
	return batches, nil
}

func (r *batchRepo) FindByProductIDFEFO(ctx context.Context, productID primitive.ObjectID) ([]model.Batch, error) {
	now := time.Now()
	filter := bson.M{
		"productId":  productID,
		"quantity":   bson.M{"$gt": 0},
		"expiryDate": bson.M{"$gt": now},
	}
	opts := options.Find().SetSort(bson.D{{Key: "expiryDate", Value: 1}})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var batches []model.Batch
	if err := cursor.All(ctx, &batches); err != nil {
		return nil, err
	}
	return batches, nil
}

func (r *batchRepo) FindExpiringBatches(ctx context.Context, clientID string, daysAhead int) ([]model.Batch, error) {
	now := time.Now()
	future := now.AddDate(0, 0, daysAhead)
	filter := bson.M{
		"clientId": clientID,
		"quantity": bson.M{"$gt": 0},
		"expiryDate": bson.M{
			"$gt":  now,
			"$lte": future,
		},
	}
	opts := options.Find().SetSort(bson.D{{Key: "expiryDate", Value: 1}})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var batches []model.Batch
	if err := cursor.All(ctx, &batches); err != nil {
		return nil, err
	}
	return batches, nil
}

func (r *batchRepo) FindLowStock(ctx context.Context, clientID string) ([]model.Batch, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"clientId": clientID, "quantity": bson.M{"$gt": 0}}}},
		{{Key: "$group", Value: bson.M{
			"_id":        "$productId",
			"totalStock": bson.M{"$sum": "$quantity"},
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "products",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "product",
		}}},
		{{Key: "$unwind", Value: "$product"}},
		{{Key: "$match", Value: bson.M{
			"$expr": bson.M{"$lte": bson.A{"$totalStock", "$product.minStock"}},
		}}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.Batch
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *batchRepo) UpdateQuantity(ctx context.Context, id primitive.ObjectID, quantity int) error {
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"quantity": quantity, "updatedDate": time.Now()}})
	return err
}

func (r *batchRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
