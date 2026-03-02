package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Allergy struct {
	DrugName string `json:"drugName" bson:"drugName"`
	Reaction string `json:"reaction" bson:"reaction"`
	Severity string `json:"severity" bson:"severity"` // MILD, MODERATE, SEVERE
}

type Patient struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	IDCard          string             `json:"idCard" bson:"idCard"`
	FirstName       string             `json:"firstName" bson:"firstName"`
	LastName        string             `json:"lastName" bson:"lastName"`
	Phone           string             `json:"phone" bson:"phone"`
	Email           string             `json:"email" bson:"email"`
	DateOfBirth     time.Time          `json:"dateOfBirth" bson:"dateOfBirth"`
	Gender          string             `json:"gender" bson:"gender"`
	Address         string             `json:"address" bson:"address"`
	Allergies       []Allergy          `json:"allergies" bson:"allergies"`
	ChronicDiseases []string           `json:"chronicDiseases" bson:"chronicDiseases"`
	Notes           string             `json:"notes" bson:"notes"`
	PDPAConsent     bool               `json:"pdpaConsent" bson:"pdpaConsent"`
	PDPAConsentDate time.Time          `json:"pdpaConsentDate" bson:"pdpaConsentDate"`
	CreatedBy       string             `json:"createdBy" bson:"createdBy"`
	CreatedDate     time.Time          `json:"createdDate" bson:"createdDate"`
	UpdatedBy       string             `json:"updatedBy" bson:"updatedBy"`
	UpdatedDate     time.Time          `json:"updatedDate" bson:"updatedDate"`
}
