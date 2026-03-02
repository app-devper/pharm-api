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

// ข.ย.9 — บัญชีซื้อยา (Purchase/Receive ledger)
func (h *ReportHandler) GetKY9(ctx *gin.Context) {
	h.getReceiveReport(ctx, "ky9")
}

// ข.ย.10 — บัญชีขายยาอันตราย
func (h *ReportHandler) GetKY10(ctx *gin.Context) {
	h.getSaleReport(ctx, "ky10")
}

// ข.ย.11 — บัญชีขายยาควบคุมพิเศษ
func (h *ReportHandler) GetKY11(ctx *gin.Context) {
	h.getSaleReport(ctx, "ky11")
}

// ข.ย.12 — บัญชีรับ-จ่ายวัตถุออกฤทธิ์ฯ
func (h *ReportHandler) GetKY12(ctx *gin.Context) {
	h.getSaleReport(ctx, "ky12")
}

// ข.ย.13 — บัญชีรับ-จ่ายยาเสพติดให้โทษ ประเภท 3
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
	case "ky11":
		w.Write([]string{"วันที่", "เลขที่ใบเสร็จ", "ชื่อยา", "จำนวน", "หน่วย", "ชื่อผู้ซื้อ", "ชื่อผู้สั่งจ่าย", "เภสัชกร"})
	case "ky12":
		w.Write([]string{"วันที่", "เลขที่ใบเสร็จ", "ชื่อยา", "จำนวน", "หน่วย", "ชื่อผู้ซื้อ", "เลขบัตรประชาชน", "เภสัชกร"})
	default:
		w.Write([]string{"วันที่", "เลขที่ใบเสร็จ", "ชื่อยา", "จำนวน", "หน่วย", "เภสัชกร"})
	}

	for _, sale := range sales {
		for _, item := range sale.Items {
			row := []string{
				sale.CreatedDate.Format("02/01/2006"),
				sale.ReceiptNumber,
				item.TradeName,
				fmt.Sprintf("%d", item.Quantity),
				item.Unit,
			}
			switch reportKey {
			case "ky11":
				row = append(row, sale.BuyerName, sale.PrescriberName, sale.PharmacistName)
			case "ky12":
				row = append(row, sale.BuyerName, sale.BuyerIDCard, sale.PharmacistName)
			default:
				row = append(row, sale.PharmacistName)
			}
			w.Write(row)
		}
	}
}

// getReceiveReport returns batches received for products that have the given reportKey in their reportTypes
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

	w.Write([]string{"วันที่รับ", "ชื่อการค้า", "ชื่อสามัญ", "Lot Number", "จำนวน", "หน่วย", "ราคาทุน", "แหล่งที่ซื้อ", "วันหมดอายุ"})
	for _, r := range rows {
		w.Write([]string{
			r.ReceivedAt.Format("02/01/2006"),
			r.TradeName,
			r.GenericName,
			r.LotNumber,
			fmt.Sprintf("%d", r.Quantity),
			r.Unit,
			fmt.Sprintf("%.2f", r.CostPrice),
			r.SupplierName,
			r.ExpiryDate.Format("02/01/2006"),
		})
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
