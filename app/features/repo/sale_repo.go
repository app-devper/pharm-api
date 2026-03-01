package repo

import (
	"context"
	"fmt"
	"pharmacy-pos/api/app/domain/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type saleRepo struct {
	col *mongo.Collection
	db  *mongo.Database
}

func NewSaleRepo(db *mongo.Database) *saleRepo {
	return &saleRepo{col: db.Collection("sales"), db: db}
}

func (r *saleRepo) Create(ctx context.Context, sale *model.Sale) (*model.Sale, error) {
	if sale.ReceiptNumber == "" {
		sale.ReceiptNumber = r.generateReceiptNumber(ctx)
	}
	result, err := r.col.InsertOne(ctx, sale)
	if err != nil {
		return nil, err
	}
	sale.ID = result.InsertedID.(primitive.ObjectID)
	return sale, nil
}

func (r *saleRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Sale, error) {
	var sale model.Sale
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&sale)
	if err != nil {
		return nil, err
	}
	return &sale, nil
}

func (r *saleRepo) FindByClientID(ctx context.Context, clientID string, page int, limit int) ([]model.Sale, int64, error) {
	filter := bson.M{"clientId": clientID}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().SetSkip(skip).SetLimit(int64(limit)).SetSort(bson.D{{Key: "createdDate", Value: -1}})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var sales []model.Sale
	if err := cursor.All(ctx, &sales); err != nil {
		return nil, 0, err
	}
	return sales, total, nil
}

func (r *saleRepo) FindByPatientID(ctx context.Context, patientID primitive.ObjectID, page int, limit int) ([]model.Sale, int64, error) {
	filter := bson.M{"patientId": patientID}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().SetSkip(skip).SetLimit(int64(limit)).SetSort(bson.D{{Key: "createdDate", Value: -1}})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var sales []model.Sale
	if err := cursor.All(ctx, &sales); err != nil {
		return nil, 0, err
	}
	return sales, total, nil
}

func (r *saleRepo) FindByDateRange(ctx context.Context, clientID string, from time.Time, to time.Time) ([]model.Sale, error) {
	filter := bson.M{
		"clientId": clientID,
		"createdDate": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}
	opts := options.Find().SetSort(bson.D{{Key: "createdDate", Value: -1}})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sales []model.Sale
	if err := cursor.All(ctx, &sales); err != nil {
		return nil, err
	}
	return sales, nil
}

func (r *saleRepo) FindControlledSales(ctx context.Context, clientID string, classification model.DrugClassification, from time.Time, to time.Time) ([]model.Sale, error) {
	// Get product IDs matching the specific classification
	productCursor, err := r.db.Collection("products").Find(ctx, bson.M{
		"clientId":           clientID,
		"drugClassification": string(classification),
	}, options.Find().SetProjection(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}
	defer productCursor.Close(ctx)

	var products []bson.M
	if err := productCursor.All(ctx, &products); err != nil {
		return nil, err
	}

	if len(products) == 0 {
		return []model.Sale{}, nil
	}

	var productIDs []interface{}
	for _, p := range products {
		productIDs = append(productIDs, p["_id"])
	}

	filter := bson.M{
		"clientId":        clientID,
		"createdDate":     bson.M{"$gte": from, "$lte": to},
		"items.productId": bson.M{"$in": productIDs},
	}
	opts := options.Find().SetSort(bson.D{{Key: "createdDate", Value: -1}})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sales []model.Sale
	if err := cursor.All(ctx, &sales); err != nil {
		return nil, err
	}
	return sales, nil
}

func (r *saleRepo) generateReceiptNumber(ctx context.Context) string {
	today := time.Now().Format("20060102")
	key := "RCP-" + today

	counterCol := r.db.Collection("counters")
	filter := bson.M{"_id": key}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result bson.M
	err := counterCol.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		// Fallback to count-based if atomic counter fails
		prefix := key + "-"
		count, _ := r.col.CountDocuments(ctx, bson.M{"receiptNumber": bson.M{"$regex": "^" + prefix}})
		return fmt.Sprintf("%s-%04d", key, count+1)
	}

	seq := int64(1)
	if v, ok := result["seq"].(int32); ok {
		seq = int64(v)
	} else if v, ok := result["seq"].(int64); ok {
		seq = v
	}
	return fmt.Sprintf("%s-%04d", key, seq)
}
