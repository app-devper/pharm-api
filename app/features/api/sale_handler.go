package api

import (
	"net/http"
	"pharmacy-pos/api/app/core/errs"
	"pharmacy-pos/api/app/domain/model"
	"pharmacy-pos/api/app/domain/usecase"
	"pharmacy-pos/api/app/features/request"
	"pharmacy-pos/api/middlewares"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SaleHandler struct {
	uc *usecase.SaleUsecase
}

func NewSaleHandler(uc *usecase.SaleUsecase) *SaleHandler {
	return &SaleHandler{uc: uc}
}

func (h *SaleHandler) Create(ctx *gin.Context) {
	var req request.CreateSaleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, err.Error()))
		return
	}

	userID := ctx.GetString(middlewares.SessionId)

	var patientID primitive.ObjectID
	if req.PatientID != "" {
		oid, err := primitive.ObjectIDFromHex(req.PatientID)
		if err != nil {
			errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, "invalid patient ID"))
			return
		}
		patientID = oid
	}

	items := make([]model.SaleItem, len(req.Items))
	for i, item := range req.Items {
		productOID, err := primitive.ObjectIDFromHex(item.ProductID)
		if err != nil {
			errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, "invalid product ID"))
			return
		}
		var batchOID primitive.ObjectID
		if item.BatchID != "" {
			batchOID, err = primitive.ObjectIDFromHex(item.BatchID)
			if err != nil {
				errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, "invalid batch ID"))
				return
			}
		}
		items[i] = model.SaleItem{
			ProductID: productOID,
			BatchID:   batchOID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Discount:  item.Discount,
		}
	}

	pm := model.PaymentMethod(req.PaymentMethod)
	if pm != model.PaymentCash && pm != model.PaymentTransfer && pm != model.PaymentCredit {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, "invalid payment method, must be CASH, TRANSFER, or CREDIT_CARD"))
		return
	}

	sale := &model.Sale{
		PatientID:      patientID,
		Items:          items,
		Discount:       req.Discount,
		PaymentMethod:  pm,
		AmountPaid:     req.AmountPaid,
		PharmacistID:   req.PharmacistID,
		PharmacistName: req.PharmacistName,
		BuyerName:      req.BuyerName,
		BuyerIDCard:    req.BuyerIDCard,
		PrescriberName: req.PrescriberName,
		Notes:          req.Notes,
	}

	result, err := h.uc.CreateSale(ctx, sale, userID)
	if err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, result)
}

func (h *SaleHandler) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	sale, err := h.uc.GetByID(ctx, id)
	if err != nil {
		errs.Response(ctx, http.StatusNotFound, errs.New(errs.ErrNotFound, "sale not found"))
		return
	}
	ctx.JSON(http.StatusOK, sale)
}

func (h *SaleHandler) GetAll(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

	sales, total, err := h.uc.GetAll(ctx, page, limit)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  sales,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *SaleHandler) CheckInteractions(ctx *gin.Context) {
	var req request.CheckInteractionsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, err.Error()))
		return
	}

	warnings, err := h.uc.CheckInteractions(ctx, req.ProductIDs)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"warnings":    warnings,
		"hasWarnings": len(warnings) > 0,
	})
}

func (h *SaleHandler) CheckAllergies(ctx *gin.Context) {
	var req request.CheckAllergiesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, err.Error()))
		return
	}

	warnings, err := h.uc.CheckAllergies(ctx, req.PatientID, req.ProductIDs)
	if err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"warnings":    warnings,
		"hasWarnings": len(warnings) > 0,
	})
}

func (h *SaleHandler) GetPatientHistory(ctx *gin.Context) {
	patientID := ctx.Param("id")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

	sales, total, err := h.uc.GetByPatientID(ctx, patientID, page, limit)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  sales,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
