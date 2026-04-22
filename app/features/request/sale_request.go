package request

type SaleItemRequest struct {
	ProductID string  `json:"productId" binding:"required"`
	BatchID   string  `json:"batchId"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	Unit      string  `json:"unit"` // selling unit (base unit or conversion unit)
	UnitPrice float64 `json:"unitPrice"`
	Discount  float64 `json:"discount"`
}

type CreateSaleRequest struct {
	PatientID           string            `json:"patientId"`
	Items               []SaleItemRequest `json:"items" binding:"required,min=1"`
	Discount            float64           `json:"discount"`
	PaymentMethod       string            `json:"paymentMethod" binding:"required"`
	AmountPaid          float64           `json:"amountPaid"`
	PharmacistID        string            `json:"pharmacistId"`
	PharmacistName      string            `json:"pharmacistName"`
	BuyerName           string            `json:"buyerName"`
	BuyerIDCard         string            `json:"buyerIdCard"`
	BuyerAddress        string            `json:"buyerAddress"` // For ข.ย. 12 & 13
	BuyerAge            string            `json:"buyerAge"`     // For ข.ย. 12
	BuyerLicense        string            `json:"buyerLicense"` // For ข.ย. 13
	PrescriberName      string            `json:"prescriberName"`
	PrescriberWorkplace string            `json:"prescriberWorkplace"` // For ข.ย. 12
	PrescriptionNo      string            `json:"prescriptionNo"`      // For ข.ย. 12
	DrugRegistration    string            `json:"drugRegistration"`    // For ข.ย. 13
	Notes               string            `json:"notes"`
}

type CheckInteractionsRequest struct {
	ProductIDs []string `json:"productIds" binding:"required,min=2"`
}

type CheckAllergiesRequest struct {
	PatientID  string   `json:"patientId" binding:"required"`
	ProductIDs []string `json:"productIds" binding:"required,min=1"`
}
