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
)

type ProductHandler struct {
	uc *usecase.ProductUsecase
}

func NewProductHandler(uc *usecase.ProductUsecase) *ProductHandler {
	return &ProductHandler{uc: uc}
}

func (h *ProductHandler) Create(ctx *gin.Context) {
	var req request.CreateProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, err.Error()))
		return
	}

	userID := ctx.GetString(middlewares.SessionId)

	product := &model.Product{
		Barcode:            req.Barcode,
		TradeName:          req.TradeName,
		GenericName:        req.GenericName,
		DrugClassification: model.DrugClassification(req.DrugClassification),
		Category:           req.Category,
		Dosage:             req.Dosage,
		Unit:               req.Unit,
		CostPrice:          req.CostPrice,
		SellingPrice:       req.SellingPrice,
		MinStock:           req.MinStock,
		Description:        req.Description,
		SideEffects:        req.SideEffects,
		Contraindications:  req.Contraindications,
		StorageCondition:   req.StorageCondition,
		Interactions:       req.Interactions,
		ReportTypes:        req.ReportTypes,
	}

	result, err := h.uc.Create(ctx, product, userID)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, result)
}

func (h *ProductHandler) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	product, err := h.uc.GetByID(ctx, id)
	if err != nil {
		errs.Response(ctx, http.StatusNotFound, errs.New(errs.ErrNotFound, "product not found"))
		return
	}
	ctx.JSON(http.StatusOK, product)
}

func (h *ProductHandler) GetAll(ctx *gin.Context) {
	search := ctx.Query("search")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

	products, total, err := h.uc.GetAll(ctx, search, page, limit)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  products,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *ProductHandler) GetByBarcode(ctx *gin.Context) {
	barcode := ctx.Param("barcode")

	product, err := h.uc.GetByBarcode(ctx, barcode)
	if err != nil {
		errs.Response(ctx, http.StatusNotFound, errs.New(errs.ErrNotFound, "product not found"))
		return
	}
	ctx.JSON(http.StatusOK, product)
}

func (h *ProductHandler) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var req request.UpdateProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, err.Error()))
		return
	}

	existing, err := h.uc.GetByID(ctx, id)
	if err != nil {
		errs.Response(ctx, http.StatusNotFound, errs.New(errs.ErrNotFound, "product not found"))
		return
	}

	userID := ctx.GetString(middlewares.SessionId)

	if req.TradeName != "" {
		existing.TradeName = req.TradeName
	}
	if req.GenericName != "" {
		existing.GenericName = req.GenericName
	}
	if req.DrugClassification != "" {
		existing.DrugClassification = model.DrugClassification(req.DrugClassification)
	}
	if req.Category != "" {
		existing.Category = req.Category
	}
	if req.Dosage != "" {
		existing.Dosage = req.Dosage
	}
	if req.Unit != "" {
		existing.Unit = req.Unit
	}
	if req.CostPrice != nil {
		existing.CostPrice = *req.CostPrice
	}
	if req.SellingPrice != nil {
		existing.SellingPrice = *req.SellingPrice
	}
	if req.MinStock != nil {
		existing.MinStock = *req.MinStock
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.SideEffects != "" {
		existing.SideEffects = req.SideEffects
	}
	if req.Contraindications != "" {
		existing.Contraindications = req.Contraindications
	}
	if req.StorageCondition != "" {
		existing.StorageCondition = req.StorageCondition
	}
	if req.Interactions != nil {
		existing.Interactions = req.Interactions
	}
	if req.ReportTypes != nil {
		existing.ReportTypes = req.ReportTypes
	}

	if err := h.uc.Update(ctx, existing, userID); err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, existing)
}

func (h *ProductHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := h.uc.Delete(ctx, id); err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
