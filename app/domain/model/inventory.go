package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Batch struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ProductID    primitive.ObjectID `json:"productId" bson:"productId"`
	LotNumber    string             `json:"lotNumber" bson:"lotNumber"`
	ExpiryDate   time.Time          `json:"expiryDate" bson:"expiryDate"`
	Quantity     int                `json:"quantity" bson:"quantity"`
	CostPrice    float64            `json:"costPrice" bson:"costPrice"`
	SupplierID   primitive.ObjectID `json:"supplierId" bson:"supplierId"`
	SupplierName string             `json:"supplierName" bson:"supplierName"`
	ReceivedAt   time.Time          `json:"receivedAt" bson:"receivedAt"`
	CreatedBy    string             `json:"createdBy" bson:"createdBy"`
	CreatedDate  time.Time          `json:"createdDate" bson:"createdDate"`
	UpdatedBy    string             `json:"updatedBy" bson:"updatedBy"`
	UpdatedDate  time.Time          `json:"updatedDate" bson:"updatedDate"`
}

type GoodsReceiptItem struct {
	ProductID  primitive.ObjectID `json:"productId" bson:"productId"`
	TradeName  string             `json:"tradeName" bson:"tradeName"`
	LotNumber  string             `json:"lotNumber" bson:"lotNumber"`
	ExpiryDate time.Time          `json:"expiryDate" bson:"expiryDate"`
	Quantity   int                `json:"quantity" bson:"quantity"`
	CostPrice  float64            `json:"costPrice" bson:"costPrice"`
}

type GoodsReceipt struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ReceiptNumber string             `json:"receiptNumber" bson:"receiptNumber"`
	SupplierName  string             `json:"supplierName" bson:"supplierName"`
	Items         []GoodsReceiptItem `json:"items" bson:"items"`
	TotalItems    int                `json:"totalItems" bson:"totalItems"`
	TotalCost     float64            `json:"totalCost" bson:"totalCost"`
	Notes         string             `json:"notes" bson:"notes"`
	CreatedBy     string             `json:"createdBy" bson:"createdBy"`
	CreatedDate   time.Time          `json:"createdDate" bson:"createdDate"`
}

type StockMovement struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ProductID   primitive.ObjectID `json:"productId" bson:"productId"`
	BatchID     primitive.ObjectID `json:"batchId" bson:"batchId"`
	Type        string             `json:"type" bson:"type"` // IN, OUT, ADJUST
	Quantity    int                `json:"quantity" bson:"quantity"`
	Reference   string             `json:"reference" bson:"reference"`
	Reason      string             `json:"reason" bson:"reason"`
	CreatedBy   string             `json:"createdBy" bson:"createdBy"`
	CreatedDate time.Time          `json:"createdDate" bson:"createdDate"`
}
