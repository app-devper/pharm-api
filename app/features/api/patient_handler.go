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
)

type PatientHandler struct {
	uc *usecase.PatientUsecase
}

func NewPatientHandler(uc *usecase.PatientUsecase) *PatientHandler {
	return &PatientHandler{uc: uc}
}

func (h *PatientHandler) Create(ctx *gin.Context) {
	var req request.CreatePatientRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, err.Error()))
		return
	}

	userID := ctx.GetString(middlewares.SessionId)

	var dob time.Time
	if req.DateOfBirth != "" {
		parsed, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, "invalid date format, use YYYY-MM-DD"))
			return
		}
		dob = parsed
	}

	allergies := make([]model.Allergy, len(req.Allergies))
	for i, a := range req.Allergies {
		allergies[i] = model.Allergy{
			DrugName: a.DrugName,
			Reaction: a.Reaction,
			Severity: a.Severity,
		}
	}

	var consentDate time.Time
	if req.PDPAConsent {
		consentDate = time.Now()
	}

	patient := &model.Patient{
		IDCard:          req.IDCard,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Phone:           req.Phone,
		Email:           req.Email,
		DateOfBirth:     dob,
		Gender:          req.Gender,
		Address:         req.Address,
		Allergies:       allergies,
		ChronicDiseases: req.ChronicDiseases,
		Notes:           req.Notes,
		PDPAConsent:     req.PDPAConsent,
		PDPAConsentDate: consentDate,
	}

	result, err := h.uc.Create(ctx, patient, userID)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, result)
}

func (h *PatientHandler) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	patient, err := h.uc.GetByID(ctx, id)
	if err != nil {
		errs.Response(ctx, http.StatusNotFound, errs.New(errs.ErrNotFound, "patient not found"))
		return
	}
	ctx.JSON(http.StatusOK, patient)
}

func (h *PatientHandler) GetAll(ctx *gin.Context) {
	search := ctx.Query("search")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

	patients, total, err := h.uc.GetAll(ctx, search, page, limit)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  patients,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *PatientHandler) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var req request.UpdatePatientRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, err.Error()))
		return
	}

	existing, err := h.uc.GetByID(ctx, id)
	if err != nil {
		errs.Response(ctx, http.StatusNotFound, errs.New(errs.ErrNotFound, "patient not found"))
		return
	}

	userID := ctx.GetString(middlewares.SessionId)

	if req.FirstName != "" {
		existing.FirstName = req.FirstName
	}
	if req.LastName != "" {
		existing.LastName = req.LastName
	}
	if req.Phone != "" {
		existing.Phone = req.Phone
	}
	if req.Email != "" {
		existing.Email = req.Email
	}
	if req.Address != "" {
		existing.Address = req.Address
	}
	if req.Allergies != nil {
		allergies := make([]model.Allergy, len(req.Allergies))
		for i, a := range req.Allergies {
			allergies[i] = model.Allergy{
				DrugName: a.DrugName,
				Reaction: a.Reaction,
				Severity: a.Severity,
			}
		}
		existing.Allergies = allergies
	}
	if req.ChronicDiseases != nil {
		existing.ChronicDiseases = req.ChronicDiseases
	}
	if req.DateOfBirth != "" {
		parsed, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, "invalid date format, use YYYY-MM-DD"))
			return
		}
		existing.DateOfBirth = parsed
	}
	if req.Gender != "" {
		existing.Gender = req.Gender
	}
	if req.Notes != "" {
		existing.Notes = req.Notes
	}

	if err := h.uc.Update(ctx, existing, userID); err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, existing)
}

func (h *PatientHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := h.uc.Delete(ctx, id); err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
