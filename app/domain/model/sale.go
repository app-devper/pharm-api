package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PaymentMethod string

const (
	PaymentCash     PaymentMethod = "CASH"
	PaymentTransfer PaymentMethod = "TRANSFER"
	PaymentCredit   PaymentMethod = "CREDIT_CARD"
)

type SaleItem struct {
	ProductID  primitive.ObjectID `json:"productId" bson:"productId"`
	BatchID    primitive.ObjectID `json:"batchId" bson:"batchId"`
	TradeName  string             `json:"tradeName" bson:"tradeName"`
	LotNumber  string             `json:"lotNumber" bson:"lotNumber"`
	Quantity   int                `json:"quantity" bson:"quantity"`
	Unit       string             `json:"unit" bson:"unit"`
	UnitPrice  float64            `json:"unitPrice" bson:"unitPrice"`
	Discount   float64            `json:"discount" bson:"discount"`
	TotalPrice float64            `json:"totalPrice" bson:"totalPrice"`
}

type Sale struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ReceiptNumber  string             `json:"receiptNumber" bson:"receiptNumber"`
	PatientID      primitive.ObjectID `json:"patientId,omitempty" bson:"patientId,omitempty"`
	Items          []SaleItem         `json:"items" bson:"items"`
	SubTotal       float64            `json:"subTotal" bson:"subTotal"`
	Discount       float64            `json:"discount" bson:"discount"`
	Total          float64            `json:"total" bson:"total"`
	PaymentMethod  PaymentMethod      `json:"paymentMethod" bson:"paymentMethod"`
	AmountPaid     float64            `json:"amountPaid" bson:"amountPaid"`
	Change         float64            `json:"change" bson:"change"`
	PharmacistID   string             `json:"pharmacistId" bson:"pharmacistId"`
	PharmacistName string             `json:"pharmacistName" bson:"pharmacistName"`
	Notes          string             `json:"notes" bson:"notes"`
	HasControlled  bool               `json:"hasControlled" bson:"hasControlled"`
	BuyerName      string             `json:"buyerName,omitempty" bson:"buyerName,omitempty"`
	BuyerIDCard    string             `json:"buyerIdCard,omitempty" bson:"buyerIdCard,omitempty"`
	PrescriberName string             `json:"prescriberName,omitempty" bson:"prescriberName,omitempty"`
	CreatedBy      string             `json:"createdBy" bson:"createdBy"`
	CreatedDate    time.Time          `json:"createdDate" bson:"createdDate"`
}
