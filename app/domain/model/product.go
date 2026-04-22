package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UnitConversion struct {
	UnitName         string  `json:"unitName" bson:"unitName"`
	ConversionFactor int     `json:"conversionFactor" bson:"conversionFactor"`
	Barcode          string  `json:"barcode" bson:"barcode"`
	SellingPrice     float64 `json:"sellingPrice" bson:"sellingPrice"`
	CostPrice        float64 `json:"costPrice" bson:"costPrice"`
}

type Product struct {
	ID                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Barcode           string             `json:"barcode" bson:"barcode"`
	TradeName         string             `json:"tradeName" bson:"tradeName"`
	GenericName       string             `json:"genericName" bson:"genericName"`
	Category          string             `json:"category" bson:"category"`
	Dosage            string             `json:"dosage" bson:"dosage"`
	Unit              string             `json:"unit" bson:"unit"`
	CostPrice         float64            `json:"costPrice" bson:"costPrice"`
	SellingPrice      float64            `json:"sellingPrice" bson:"sellingPrice"`
	MinStock          int                `json:"minStock" bson:"minStock"`
	Description       string             `json:"description" bson:"description"`
	SideEffects       string             `json:"sideEffects" bson:"sideEffects"`
	Contraindications string             `json:"contraindications" bson:"contraindications"`
	StorageCondition  string             `json:"storageCondition" bson:"storageCondition"`
	UnitConversions   []UnitConversion   `json:"unitConversions" bson:"unitConversions"`
	Interactions      []string           `json:"interactions" bson:"interactions"`
	ReportTypes       []string           `json:"reportTypes" bson:"reportTypes"`
	Status            string             `json:"status" bson:"status"`
	CreatedBy         string             `json:"createdBy" bson:"createdBy"`
	CreatedDate       time.Time          `json:"createdDate" bson:"createdDate"`
	UpdatedBy         string             `json:"updatedBy" bson:"updatedBy"`
	UpdatedDate       time.Time          `json:"updatedDate" bson:"updatedDate"`
}
