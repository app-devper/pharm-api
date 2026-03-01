package api

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"pharmacy-pos/api/app/core/errs"
	"pharmacy-pos/api/app/domain/model"
	"pharmacy-pos/api/middlewares"
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

func (h *ReportHandler) GetKY9(ctx *gin.Context) {
	h.getControlledReport(ctx, model.DrugDangerous, "ky9")
}

func (h *ReportHandler) GetKY10(ctx *gin.Context) {
	h.getControlledReport(ctx, model.DrugControlled, "ky10")
}

func (h *ReportHandler) GetKY11(ctx *gin.Context) {
	h.getControlledReport(ctx, model.DrugPsycho, "ky11")
}

func (h *ReportHandler) GetKY12(ctx *gin.Context) {
	h.getControlledReport(ctx, model.DrugNarcotic, "ky12")
}

type reportSale struct {
	model.Sale `bson:",inline"`
}

func (h *ReportHandler) getControlledReport(ctx *gin.Context, classification model.DrugClassification, reportName string) {
	clientID := ctx.GetString(middlewares.ClientId)
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

	// Get all product IDs with the given classification
	productFilter := bson.M{"clientId": clientID, "drugClassification": string(classification)}
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
			h.writeEmptyCSV(ctx, reportName)
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
		"clientId":        clientID,
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

	// Filter items to only include items matching the classification
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
		h.writeCSV(ctx, sales, classification, reportName)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": sales,
		"from": fromStr,
		"to":   toStr,
	})
}

func (h *ReportHandler) writeCSV(ctx *gin.Context, sales []model.Sale, classification model.DrugClassification, reportName string) {
	ctx.Header("Content-Type", "text/csv; charset=utf-8")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s.csv", reportName, time.Now().Format("20060102")))
	ctx.Writer.Write([]byte{0xEF, 0xBB, 0xBF}) // UTF-8 BOM

	w := csv.NewWriter(ctx.Writer)
	defer w.Flush()

	switch classification {
	case model.DrugControlled:
		w.Write([]string{"วันที่", "เลขที่ใบเสร็จ", "ชื่อยา", "จำนวน", "หน่วย", "ชื่อผู้ซื้อ", "ชื่อผู้สั่งจ่าย", "เภสัชกร"})
	case model.DrugPsycho:
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
			switch classification {
			case model.DrugControlled:
				row = append(row, sale.BuyerName, sale.PrescriberName, sale.PharmacistName)
			case model.DrugPsycho:
				row = append(row, sale.BuyerName, sale.BuyerIDCard, sale.PharmacistName)
			default:
				row = append(row, sale.PharmacistName)
			}
			w.Write(row)
		}
	}
}

func (h *ReportHandler) writeEmptyCSV(ctx *gin.Context, reportName string) {
	ctx.Header("Content-Type", "text/csv; charset=utf-8")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s.csv", reportName, time.Now().Format("20060102")))
	ctx.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(ctx.Writer)
	w.Write([]string{"ไม่มีข้อมูล"})
	w.Flush()
}
