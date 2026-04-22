package api

import (
	"context"
	"fmt"
	"net/http"
	"pharmacy-pos/api/app/core/errs"
	"pharmacy-pos/api/app/domain/model"
	"pharmacy-pos/api/app/domain/usecase"
	"pharmacy-pos/api/app/features/request"
	"pharmacy-pos/api/middlewares"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ReceiveHandler struct {
	db      *mongo.Database
	batchUC *usecase.BatchUsecase
}

func NewReceiveHandler(db *mongo.Database, batchUC *usecase.BatchUsecase) *ReceiveHandler {
	return &ReceiveHandler{db: db, batchUC: batchUC}
}

func (h *ReceiveHandler) Create(ctx *gin.Context) {
	var req request.CreateGoodsReceiptRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, err.Error()))
		return
	}

	userID := ctx.GetString(middlewares.SessionId)
	c := ctx.Request.Context()

	// Phase 1: Validate all items and build data (no DB writes)
	type validatedItem struct {
		receiptItem model.GoodsReceiptItem
		batch       *model.Batch
	}

	var validated []validatedItem
	var totalCost float64

	for i, item := range req.Items {
		productOID, err := primitive.ObjectIDFromHex(item.ProductID)
		if err != nil {
			errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, fmt.Sprintf("item %d: invalid product ID", i+1)))
			return
		}

		expiryDate, err := time.Parse("2006-01-02", item.ExpiryDate)
		if err != nil {
			errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, fmt.Sprintf("item %d: invalid expiry date, use YYYY-MM-DD", i+1)))
			return
		}

		// Lookup product for trade name and unit conversions
		var product model.Product
		err = h.db.Collection("products").FindOne(c, bson.M{"_id": productOID}).Decode(&product)
		tradeName := ""
		if err == nil {
			tradeName = product.TradeName
		}

		// Apply unit conversion: convert to base unit quantity
		baseQty := item.Quantity
		baseCost := item.CostPrice
		if item.ReceiveUnit != "" && item.ReceiveUnit != product.Unit {
			found := false
			for _, uc := range product.UnitConversions {
				if uc.UnitName == item.ReceiveUnit {
					baseQty = item.Quantity * uc.ConversionFactor
					if uc.ConversionFactor > 0 {
						baseCost = item.CostPrice / float64(uc.ConversionFactor)
					}
					found = true
					break
				}
			}
			if !found {
				errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, fmt.Sprintf("item %d: unknown receive unit '%s'", i+1, item.ReceiveUnit)))
				return
			}
		}

		validated = append(validated, validatedItem{
			receiptItem: model.GoodsReceiptItem{
				ProductID:  productOID,
				TradeName:  tradeName,
				LotNumber:  item.LotNumber,
				ExpiryDate: expiryDate,
				Quantity:   item.Quantity,
				CostPrice:  item.CostPrice,
			},
			batch: &model.Batch{
				ProductID:    productOID,
				LotNumber:    item.LotNumber,
				ExpiryDate:   expiryDate,
				Quantity:     baseQty,
				CostPrice:    baseCost,
				SupplierName: req.SupplierName,
			},
		})

		totalCost += item.CostPrice * float64(item.Quantity)
	}

	// Phase 2: All validated — now write batches and receipt in a transaction
	session, err := h.db.Client().StartSession()
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, "failed to start transaction"))
		return
	}
	defer session.EndSession(c)

	var receipt *model.GoodsReceipt
	result, txErr := session.WithTransaction(c, func(sessCtx mongo.SessionContext) (interface{}, error) {
		items := make([]model.GoodsReceiptItem, 0, len(validated))
		for i, v := range validated {
			_, err := h.batchUC.ReceiveGoods(sessCtx, v.batch, userID)
			if err != nil {
				return nil, fmt.Errorf("item %d: failed to create batch: %s", i+1, err.Error())
			}
			items = append(items, v.receiptItem)
		}

		// Insert goods receipt
		receipt := model.GoodsReceipt{
			ReceiptNumber: h.generateReceiptNumber(sessCtx),
			SupplierName:  req.SupplierName,
			Items:         items,
			TotalItems:    len(items),
			TotalCost:     totalCost,
			Notes:         req.Notes,
			CreatedBy:     userID,
			CreatedDate:   time.Now(),
		}

		insertResult, err := h.db.Collection("goods_receipts").InsertOne(sessCtx, receipt)
		if err != nil {
			return nil, fmt.Errorf("failed to create goods receipt: %s", err.Error())
		}
		receipt.ID = insertResult.InsertedID.(primitive.ObjectID)

		return receipt, nil
	})

	if txErr != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, txErr.Error()))
		return
	}

	receipt = result.(*model.GoodsReceipt)

	ctx.JSON(http.StatusCreated, receipt)
}

func (h *ReceiveHandler) GetAll(ctx *gin.Context) {
	c := ctx.Request.Context()
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

	col := h.db.Collection("goods_receipts")
	filter := bson.M{}

	total, err := col.CountDocuments(c, filter)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().SetSkip(skip).SetLimit(int64(limit)).SetSort(bson.D{{Key: "createdDate", Value: -1}})

	cursor, err := col.Find(c, filter, opts)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	defer cursor.Close(c)

	var receipts []model.GoodsReceipt
	if err := cursor.All(c, &receipts); err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  receipts,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *ReceiveHandler) GetByID(ctx *gin.Context) {
	c := ctx.Request.Context()
	id := ctx.Param("id")

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, "invalid ID"))
		return
	}

	var receipt model.GoodsReceipt
	err = h.db.Collection("goods_receipts").FindOne(c, bson.M{"_id": oid}).Decode(&receipt)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			errs.Response(ctx, http.StatusNotFound, errs.New(errs.ErrNotFound, "goods receipt not found"))
			return
		}
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, receipt)
}

func (h *ReceiveHandler) generateReceiptNumber(ctx context.Context) string {
	today := time.Now().Format("20060102")
	key := "GR-" + today

	counterCol := h.db.Collection("counters")
	filter := bson.M{"_id": key}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result bson.M
	err := counterCol.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		col := h.db.Collection("goods_receipts")
		prefix := key + "-"
		count, _ := col.CountDocuments(ctx, bson.M{"receiptNumber": bson.M{"$regex": "^" + prefix}})
		return fmt.Sprintf("%s-%04d", key, count+1)
	}

	seq := int64(1)
	if v, ok := result["seq"].(int32); ok {
		seq = int64(v)
	} else if v, ok := result["seq"].(int64); ok {
		seq = v
	}
	return fmt.Sprintf("%s-%04d", key, seq)
}
