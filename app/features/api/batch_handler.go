package api

import (
	"net/http"
	"pharmacy-pos/api/app/core/errs"
	"pharmacy-pos/api/app/domain/model"
	"pharmacy-pos/api/app/domain/usecase"
	"pharmacy-pos/api/app/features/request"
	"pharmacy-pos/api/middlewares"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BatchHandler struct {
	uc *usecase.BatchUsecase
}

func NewBatchHandler(uc *usecase.BatchUsecase) *BatchHandler {
	return &BatchHandler{uc: uc}
}

func (h *BatchHandler) Create(ctx *gin.Context) {
	var req request.CreateBatchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, err.Error()))
		return
	}

	clientID := ctx.GetString(middlewares.ClientId)
	userID := ctx.GetString(middlewares.SessionId)

	productOID, err := primitive.ObjectIDFromHex(req.ProductID)
	if err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, "invalid product ID"))
		return
	}

	expiryDate, err := time.Parse("2006-01-02", req.ExpiryDate)
	if err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, "invalid expiry date, use YYYY-MM-DD"))
		return
	}

	var supplierOID primitive.ObjectID
	if req.SupplierID != "" {
		supplierOID, err = primitive.ObjectIDFromHex(req.SupplierID)
		if err != nil {
			errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, "invalid supplier ID"))
			return
		}
	}

	batch := &model.Batch{
		ClientID:   clientID,
		ProductID:  productOID,
		LotNumber:  req.LotNumber,
		ExpiryDate: expiryDate,
		Quantity:   req.Quantity,
		CostPrice:  req.CostPrice,
		SupplierID: supplierOID,
	}

	result, err := h.uc.ReceiveGoods(ctx, batch, userID)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, result)
}

func (h *BatchHandler) GetByProductID(ctx *gin.Context) {
	productID := ctx.Param("productId")
	batches, err := h.uc.GetByProductID(ctx, productID)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, batches)
}

func (h *BatchHandler) GetExpiringBatches(ctx *gin.Context) {
	clientID := ctx.GetString(middlewares.ClientId)
	days, _ := strconv.Atoi(ctx.DefaultQuery("days", "180"))

	batches, err := h.uc.GetExpiringBatches(ctx, clientID, days)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, batches)
}

func (h *BatchHandler) GetLowStock(ctx *gin.Context) {
	clientID := ctx.GetString(middlewares.ClientId)

	batches, err := h.uc.GetLowStock(ctx, clientID)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, batches)
}

func (h *BatchHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := h.uc.Delete(ctx, id); err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
