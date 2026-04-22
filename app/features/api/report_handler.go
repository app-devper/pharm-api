package api

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"pharmacy-pos/api/app/core/errs"
	"pharmacy-pos/api/app/domain/model"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ReportHandler struct {
	db *mongo.Database
}

func NewReportHandler(db *mongo.Database) *ReportHandler {
	return &ReportHandler{db: db}
}

// ข.ย.9 — บัญชีการซื้อยา
func (h *ReportHandler) GetKY9(ctx *gin.Context) {
	h.getReceiveReport(ctx, "ky9")
}

// ข.ย.10 — บัญชีการขายยาควบคุมพิเศษ
func (h *ReportHandler) GetKY10(ctx *gin.Context) {
	h.getSaleReport(ctx, "ky10")
}

// ข.ย.11 — บัญชีการขายยาอันตราย
func (h *ReportHandler) GetKY11(ctx *gin.Context) {
	h.getSaleReport(ctx, "ky11")
}

// ข.ย.12 — บัญชีการขายยาตามใบสั่งของผู้ประกอบวิชาชีพฯ
func (h *ReportHandler) GetKY12(ctx *gin.Context) {
	h.getSaleReport(ctx, "ky12")
}

// ข.ย.13 — รายงานการขายยาตามที่เลขาธิการคณะกรรมการอาหารและยากำหนด
func (h *ReportHandler) GetKY13(ctx *gin.Context) {
	h.getSaleReport(ctx, "ky13")
}

type reportSale struct {
	model.Sale `bson:",inline"`
}

type receiveRow struct {
	ReceivedAt   time.Time `json:"receivedAt"`
	TradeName    string    `json:"tradeName"`
	GenericName  string    `json:"genericName"`
	LotNumber    string    `json:"lotNumber"`
	Quantity     int       `json:"quantity"`
	Unit         string    `json:"unit"`
	CostPrice    float64   `json:"costPrice"`
	SupplierName string    `json:"supplierName"`
	ExpiryDate   time.Time `json:"expiryDate"`
}

func (h *ReportHandler) getSaleReport(ctx *gin.Context, reportKey string) {
	// Note: Reports are filtered by product.reportTypes array, not by drugClassification
	// This allows flexible configuration - any product can be assigned to any report type
	c := ctx.Request.Context()

	fromStr := ctx.DefaultQuery("from", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	toStr := ctx.DefaultQuery("to", time.Now().Format("2006-01-02"))

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, "invalid from date"))
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, "invalid to date"))
		return
	}
	to = to.AddDate(0, 0, 1)

	// Get all products that have this reportKey in their reportTypes array
	productFilter := bson.M{"reportTypes": reportKey}
	productCursor, err := h.db.Collection("products").Find(c, productFilter, options.Find().SetProjection(bson.M{"_id": 1, "tradeName": 1}))
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	var products []bson.M
	productCursor.All(c, &products)
	productCursor.Close(c)

	if len(products) == 0 {
		format := ctx.Query("format")
		if format == "csv" {
			h.writeEmptyCSV(ctx, reportKey)
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"data": []interface{}{}, "from": fromStr, "to": toStr})
		return
	}

	var productIDs []interface{}
	productIDSet := make(map[primitive.ObjectID]bool)
	for _, p := range products {
		productIDs = append(productIDs, p["_id"])
		if oid, ok := p["_id"].(primitive.ObjectID); ok {
			productIDSet[oid] = true
		}
	}

	// Find sales containing these products
	salesFilter := bson.M{
		"createdDate":     bson.M{"$gte": from, "$lt": to},
		"items.productId": bson.M{"$in": productIDs},
	}
	opts := options.Find().SetSort(bson.D{{Key: "createdDate", Value: 1}})

	cursor, err := h.db.Collection("sales").Find(c, salesFilter, opts)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	defer cursor.Close(c)

	var sales []model.Sale
	if err := cursor.All(c, &sales); err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}

	// Filter items to only include items matching the products
	for i := range sales {
		var filtered []model.SaleItem
		for _, item := range sales[i].Items {
			if productIDSet[item.ProductID] {
				filtered = append(filtered, item)
			}
		}
		sales[i].Items = filtered
	}

	format := ctx.Query("format")
	if format == "csv" {
		h.writeCSV(ctx, sales, reportKey)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": sales,
		"from": fromStr,
		"to":   toStr,
	})
}

func (h *ReportHandler) writeCSV(ctx *gin.Context, sales []model.Sale, reportKey string) {
	ctx.Header("Content-Type", "text/csv; charset=utf-8")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s.csv", reportKey, time.Now().Format("20060102")))
	ctx.Writer.Write([]byte{0xEF, 0xBB, 0xBF}) // UTF-8 BOM

	w := csv.NewWriter(ctx.Writer)
	defer w.Flush()

	switch reportKey {
	case "ky10":
		w.Write([]string{"ลำดับที่", "วัน/เดือน/ปี ที่ขาย", "ชื่อยา", "เลขที่หรืออักษรของครั้งที่ผลิต (Lot No.)", "จำนวน/ปริมาณที่ขาย", "ชื่อและสกุลของผู้ซื้อ", "ลายมือชื่อผู้มีหน้าที่ปฏิบัติการ", "หมายเหตุ"})
	case "ky11":
		w.Write([]string{"ลำดับที่", "วัน/เดือน/ปี ที่ขาย", "ชื่อยา", "เลขที่หรืออักษรของครั้งที่ผลิต (Lot No.)", "จำนวน/ปริมาณที่ขาย", "ชื่อและสกุลของผู้ซื้อ", "ลายมือชื่อผู้มีหน้าที่ปฏิบัติการ", "หมายเหตุ"})
	case "ky12":
		w.Write([]string{"ลำดับที่", "วัน/เดือน/ปี ที่ขาย", "ชื่อยา", "เลขที่หรืออักษรของครั้งที่ผลิต (Lot No.)", "จำนวน/ปริมาณที่ขาย", "ชื่อ-สกุลผู้ใช้ยา", "อายุผู้ใช้ยา", "ที่อยู่ผู้ใช้ยา", "ชื่อและสกุลผู้สั่งยา", "สถานที่ทำงานของผู้สั่งยา", "ลายมือชื่อผู้มีหน้าที่ปฏิบัติการ", "หมายเหตุ"})
	case "ky13":
		w.Write([]string{"ลำดับที่", "วัน/เดือน/ปี ที่ขาย", "ชื่อยา", "เลขทะเบียนตำรับยา", "เลขที่หรืออักษรของครั้งที่ผลิต (Lot No.)", "จำนวน/ปริมาณที่ขาย", "ชื่อผู้ซื้อหรือชื่อสถานพยาบาล/ร้านยา", "ที่อยู่ผู้ซื้อ", "เลขที่ใบอนุญาตผู้ซื้อ", "หมายเหตุ"})
	default:
		w.Write([]string{"ลำดับที่", "วันที่", "ชื่อยา", "จำนวน", "หน่วย", "เภสัชกร"})
	}

	sequence := 1
	for _, sale := range sales {
		for _, item := range sale.Items {
			row := []string{
				fmt.Sprintf("%d", sequence),           // ลำดับที่
				sale.CreatedDate.Format("02/01/2006"), // วัน/เดือน/ปี
				item.TradeName,                        // ชื่อยา
			}
			switch reportKey {
			case "ky10":
				row = append(row, item.LotNumber, fmt.Sprintf("%d", item.Quantity), sale.BuyerName, sale.PharmacistName, sale.Notes)
			case "ky11":
				row = append(row, item.LotNumber, fmt.Sprintf("%d", item.Quantity), sale.BuyerName, sale.PharmacistName, sale.Notes)
			case "ky12":
				row = append(row, item.LotNumber, fmt.Sprintf("%d", item.Quantity), sale.BuyerName, sale.BuyerAge, sale.BuyerAddress, sale.PrescriberName, sale.PrescriberWorkplace, sale.PharmacistName, sale.Notes)
			case "ky13":
				row = append(row, sale.DrugRegistration, item.LotNumber, fmt.Sprintf("%d", item.Quantity), sale.BuyerName, sale.BuyerAddress, sale.BuyerLicense, sale.Notes)
			default:
				row = append(row, fmt.Sprintf("%d", item.Quantity), item.Unit, sale.PharmacistName)
			}
			w.Write(row)
			sequence++
		}
	}
}

// getReceiveReport returns batches received for products that have the given reportKey in their reportTypes
// Note: This is configuration-based, not based on drugClassification - any product can be configured for any report
func (h *ReportHandler) getReceiveReport(ctx *gin.Context, reportKey string) {
	c := ctx.Request.Context()

	fromStr := ctx.DefaultQuery("from", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	toStr := ctx.DefaultQuery("to", time.Now().Format("2006-01-02"))

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, "invalid from date"))
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, "invalid to date"))
		return
	}
	to = to.AddDate(0, 0, 1)

	// Get products that have this reportKey in their reportTypes array
	productFilter := bson.M{"reportTypes": reportKey}
	productCursor, err := h.db.Collection("products").Find(c, productFilter, options.Find().SetProjection(bson.M{"_id": 1, "tradeName": 1, "genericName": 1, "unit": 1}))
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	var products []bson.M
	productCursor.All(c, &products)
	productCursor.Close(c)

	if len(products) == 0 {
		format := ctx.Query("format")
		if format == "csv" {
			h.writeEmptyCSV(ctx, reportKey)
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"data": []interface{}{}, "from": fromStr, "to": toStr})
		return
	}

	// Build product lookup map
	type productInfo struct {
		TradeName   string
		GenericName string
		Unit        string
	}
	productMap := make(map[primitive.ObjectID]productInfo)
	var productIDs []primitive.ObjectID
	for _, p := range products {
		oid := p["_id"].(primitive.ObjectID)
		pi := productInfo{}
		if v, ok := p["tradeName"].(string); ok {
			pi.TradeName = v
		}
		if v, ok := p["genericName"].(string); ok {
			pi.GenericName = v
		}
		if v, ok := p["unit"].(string); ok {
			pi.Unit = v
		}
		productMap[oid] = pi
		productIDs = append(productIDs, oid)
	}

	// Find batches received in the date range for these products
	batchFilter := bson.M{
		"productId":  bson.M{"$in": productIDs},
		"receivedAt": bson.M{"$gte": from, "$lt": to},
	}
	opts := options.Find().SetSort(bson.D{{Key: "receivedAt", Value: 1}})

	cursor, err := h.db.Collection("batches").Find(c, batchFilter, opts)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	defer cursor.Close(c)

	var batches []model.Batch
	if err := cursor.All(c, &batches); err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}

	var rows []receiveRow
	for _, b := range batches {
		pi := productMap[b.ProductID]
		rows = append(rows, receiveRow{
			ReceivedAt:   b.ReceivedAt,
			TradeName:    pi.TradeName,
			GenericName:  pi.GenericName,
			LotNumber:    b.LotNumber,
			Quantity:     b.Quantity,
			Unit:         pi.Unit,
			CostPrice:    b.CostPrice,
			SupplierName: b.SupplierName,
			ExpiryDate:   b.ExpiryDate,
		})
	}

	format := ctx.Query("format")
	if format == "csv" {
		h.writeReceiveCSV(ctx, rows, reportKey)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": rows,
		"from": fromStr,
		"to":   toStr,
	})
}

func (h *ReportHandler) writeReceiveCSV(ctx *gin.Context, rows []receiveRow, reportKey string) {
	ctx.Header("Content-Type", "text/csv; charset=utf-8")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s.csv", reportKey, time.Now().Format("20060102")))
	ctx.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(ctx.Writer)
	defer w.Flush()

	// ข.ย. 9 - บัญชีการซื้อยา
	w.Write([]string{"ลำดับที่", "วัน/เดือน/ปี ที่ซื้อ", "ชื่อและที่อยู่ของผู้ขาย", "ชื่อยา", "เลขที่หรืออักษรของครั้งที่ผลิต (Lot No.)", "จำนวน/ปริมาณที่ซื้อ", "ลายมือชื่อผู้มีหน้าที่ปฏิบัติการ", "หมายเหตุ"})

	sequence := 1
	for _, r := range rows {
		w.Write([]string{
			fmt.Sprintf("%d", sequence),       // ลำดับที่
			r.ReceivedAt.Format("02/01/2006"), // วัน/เดือน/ปี ที่ซื้อ
			r.SupplierName,                    // ชื่อและที่อยู่ของผู้ขาย
			r.TradeName,                       // ชื่อยา
			r.LotNumber,                       // เลขที่หรืออักษรของครั้งที่ผลิต (Lot No.)
			fmt.Sprintf("%d", r.Quantity),     // จำนวน/ปริมาณที่ซื้อ
			"N/A",                             // ลายมือชื่อผู้มีหน้าที่ปฏิบัติการ (TODO: Add pharmacist signature)
			"",                                // หมายเหตุ
		})
		sequence++
	}
}

func (h *ReportHandler) writeEmptyCSV(ctx *gin.Context, reportKey string) {
	ctx.Header("Content-Type", "text/csv; charset=utf-8")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s.csv", reportKey, time.Now().Format("20060102")))
	ctx.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(ctx.Writer)
	w.Write([]string{"ไม่มีข้อมูล"})
	w.Flush()
}
