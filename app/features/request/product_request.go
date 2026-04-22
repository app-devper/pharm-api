package request

type UnitConversionRequest struct {
	UnitName         string  `json:"unitName" binding:"required"`
	ConversionFactor int     `json:"conversionFactor" binding:"required,min=1"`
	Barcode          string  `json:"barcode"`
	SellingPrice     float64 `json:"sellingPrice"`
	CostPrice        float64 `json:"costPrice"`
}

type CreateProductRequest struct {
	Barcode           string                  `json:"barcode"`
	TradeName         string                  `json:"tradeName" binding:"required"`
	GenericName       string                  `json:"genericName"`
	Category          string                  `json:"category"`
	Dosage            string                  `json:"dosage"`
	Unit              string                  `json:"unit" binding:"required"`
	CostPrice         float64                 `json:"costPrice" binding:"required"`
	SellingPrice      float64                 `json:"sellingPrice" binding:"required"`
	MinStock          int                     `json:"minStock"`
	Description       string                  `json:"description"`
	SideEffects       string                  `json:"sideEffects"`
	Contraindications string                  `json:"contraindications"`
	StorageCondition  string                  `json:"storageCondition"`
	UnitConversions   []UnitConversionRequest `json:"unitConversions"`
	Interactions      []string                `json:"interactions"`
	ReportTypes       []string                `json:"reportTypes"`
}

type UpdateProductRequest struct {
	TradeName         string                  `json:"tradeName"`
	GenericName       string                  `json:"genericName"`
	Category          string                  `json:"category"`
	Dosage            string                  `json:"dosage"`
	Unit              string                  `json:"unit"`
	CostPrice         *float64                `json:"costPrice"`
	SellingPrice      *float64                `json:"sellingPrice"`
	MinStock          *int                    `json:"minStock"`
	Description       string                  `json:"description"`
	SideEffects       string                  `json:"sideEffects"`
	Contraindications string                  `json:"contraindications"`
	StorageCondition  string                  `json:"storageCondition"`
	UnitConversions   []UnitConversionRequest `json:"unitConversions"`
	Interactions      []string                `json:"interactions"`
	ReportTypes       []string                `json:"reportTypes"`
}
