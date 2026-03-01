package request

type CreateBatchRequest struct {
	ProductID    string  `json:"productId" binding:"required"`
	LotNumber    string  `json:"lotNumber" binding:"required"`
	ExpiryDate   string  `json:"expiryDate" binding:"required"`
	Quantity     int     `json:"quantity" binding:"required,min=1"`
	CostPrice    float64 `json:"costPrice"`
	SupplierID   string  `json:"supplierId"`
	SupplierName string  `json:"supplierName"`
}
