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

type patientRepo struct {
	col *mongo.Collection
}

func NewPatientRepo(db *mongo.Database) *patientRepo {
	return &patientRepo{col: db.Collection("patients")}
}

func (r *patientRepo) Create(ctx context.Context, patient *model.Patient) (*model.Patient, error) {
	result, err := r.col.InsertOne(ctx, patient)
	if err != nil {
		return nil, err
	}
	patient.ID = result.InsertedID.(primitive.ObjectID)
	return patient, nil
}

func (r *patientRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Patient, error) {
	var patient model.Patient
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&patient)
	if err != nil {
		return nil, err
	}
	return &patient, nil
}

func (r *patientRepo) FindByIDCard(ctx context.Context, clientID string, idCard string) (*model.Patient, error) {
	var patient model.Patient
	err := r.col.FindOne(ctx, bson.M{"clientId": clientID, "idCard": idCard}).Decode(&patient)
	if err != nil {
		return nil, err
	}
	return &patient, nil
}

func (r *patientRepo) FindByClientID(ctx context.Context, clientID string, search string, page int, limit int) ([]model.Patient, int64, error) {
	filter := bson.M{"clientId": clientID}
	if search != "" {
		escaped := regexp.QuoteMeta(search)
		filter["$or"] = []bson.M{
			{"firstName": bson.M{"$regex": escaped, "$options": "i"}},
			{"lastName": bson.M{"$regex": escaped, "$options": "i"}},
			{"idCard": bson.M{"$regex": escaped, "$options": "i"}},
			{"phone": bson.M{"$regex": escaped, "$options": "i"}},
		}
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().SetSkip(skip).SetLimit(int64(limit)).SetSort(bson.D{{Key: "firstName", Value: 1}})

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var patients []model.Patient
	if err := cursor.All(ctx, &patients); err != nil {
		return nil, 0, err
	}
	return patients, total, nil
}

func (r *patientRepo) Update(ctx context.Context, patient *model.Patient) error {
	patient.UpdatedDate = time.Now()
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": patient.ID}, bson.M{"$set": patient})
	return err
}

func (r *patientRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
