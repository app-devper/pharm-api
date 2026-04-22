package request

type GoodsReceiptItemRequest struct {
	ProductID   string  `json:"productId" binding:"required"`
	LotNumber   string  `json:"lotNumber" binding:"required"`
	ExpiryDate  string  `json:"expiryDate" binding:"required"`
	Quantity    int     `json:"quantity" binding:"required,min=1"`
	CostPrice   float64 `json:"costPrice"`
	ReceiveUnit string  `json:"receiveUnit"`
}

type CreateGoodsReceiptRequest struct {
	SupplierName string                    `json:"supplierName"`
	Notes        string                    `json:"notes"`
	Items        []GoodsReceiptItemRequest `json:"items" binding:"required,min=1,dive"`
}
