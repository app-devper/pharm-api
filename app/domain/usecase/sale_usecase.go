package usecase

import (
	"context"
	"errors"
	"pharmacy-pos/api/app/domain/model"
	"pharmacy-pos/api/app/domain/repository"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type InteractionWarning struct {
	Drug1   string `json:"drug1"`
	Drug2   string `json:"drug2"`
	Message string `json:"message"`
}

type AllergyWarning struct {
	DrugName string `json:"drugName"`
	Reaction string `json:"reaction"`
	Severity string `json:"severity"`
}

type SaleUsecase struct {
	saleRepo    repository.SaleRepository
	productRepo repository.ProductRepository
	patientRepo repository.PatientRepository
	batchRepo   repository.BatchRepository
	mongoClient *mongo.Client
}

func NewSaleUsecase(
	saleRepo repository.SaleRepository,
	productRepo repository.ProductRepository,
	patientRepo repository.PatientRepository,
	batchRepo repository.BatchRepository,
	mongoClient *mongo.Client,
) *SaleUsecase {
	return &SaleUsecase{
		saleRepo:    saleRepo,
		productRepo: productRepo,
		patientRepo: patientRepo,
		batchRepo:   batchRepo,
		mongoClient: mongoClient,
	}
}

func (u *SaleUsecase) CreateSale(ctx context.Context, sale *model.Sale, userID string) (*model.Sale, error) {
	if len(sale.Items) == 0 {
		return nil, errors.New("sale must have at least one item")
	}

	hasControlled := false
	var subTotal float64
	for i, item := range sale.Items {
		product, err := u.productRepo.FindByID(ctx, item.ProductID)
		if err != nil {
			return nil, errors.New("product not found: " + item.ProductID.Hex())
		}

		sale.Items[i].TradeName = product.TradeName
		sale.Items[i].Unit = product.Unit
		if sale.Items[i].UnitPrice == 0 {
			sale.Items[i].UnitPrice = product.SellingPrice
		}
		sale.Items[i].TotalPrice = sale.Items[i].UnitPrice*float64(sale.Items[i].Quantity) - sale.Items[i].Discount

		subTotal += sale.Items[i].TotalPrice

		if product.DrugClassification == model.DrugControlled ||
			product.DrugClassification == model.DrugPsycho ||
			product.DrugClassification == model.DrugNarcotic {
			hasControlled = true
		}
	}

	sale.HasControlled = hasControlled
	sale.SubTotal = subTotal
	sale.Total = subTotal - sale.Discount
	if sale.Total < 0 {
		return nil, errors.New("total discount cannot exceed subtotal")
	}
	if sale.AmountPaid > 0 {
		if sale.AmountPaid < sale.Total {
			return nil, errors.New("amount paid is less than total")
		}
		sale.Change = sale.AmountPaid - sale.Total
	}

	if hasControlled {
		if sale.PharmacistName == "" {
			return nil, errors.New("pharmacist name is required for controlled drug sales")
		}
		if sale.BuyerName == "" {
			return nil, errors.New("buyer name is required for controlled drug sales")
		}
	}

	sale.CreatedBy = userID
	sale.CreatedDate = time.Now()

	// Use a transaction for stock deduction + sale creation
	session, err := u.mongoClient.StartSession()
	if err != nil {
		return nil, errors.New("failed to start transaction session")
	}
	defer session.EndSession(ctx)

	var result *model.Sale
	_, txErr := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// FEFO stock deduction: deduct from batches (first expired, first out)
		for i, item := range sale.Items {
			batches, err := u.batchRepo.FindByProductIDFEFO(sessCtx, item.ProductID)
			if err != nil {
				return nil, errors.New("failed to find batches for product: " + item.ProductID.Hex())
			}

			remaining := item.Quantity
			for _, batch := range batches {
				if remaining <= 0 {
					break
				}
				deduct := remaining
				if deduct > batch.Quantity {
					deduct = batch.Quantity
				}
				newQty := batch.Quantity - deduct
				if err := u.batchRepo.UpdateQuantity(sessCtx, batch.ID, newQty); err != nil {
					return nil, errors.New("failed to update batch quantity")
				}
				// Set the batch info on the first matching batch for the sale item
				if remaining == item.Quantity {
					sale.Items[i].BatchID = batch.ID
					sale.Items[i].LotNumber = batch.LotNumber
				}
				remaining -= deduct
			}
			if remaining > 0 {
				return nil, errors.New("insufficient stock for product: " + sale.Items[i].TradeName)
			}
		}

		created, err := u.saleRepo.Create(sessCtx, sale)
		if err != nil {
			return nil, err
		}
		result = created
		return nil, nil
	})
	if txErr != nil {
		return nil, txErr
	}

	return result, nil
}

func (u *SaleUsecase) GetByID(ctx context.Context, id string) (*model.Sale, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return u.saleRepo.FindByID(ctx, oid)
}

func (u *SaleUsecase) GetByClientID(ctx context.Context, clientID string, page int, limit int) ([]model.Sale, int64, error) {
	return u.saleRepo.FindByClientID(ctx, clientID, page, limit)
}

func (u *SaleUsecase) GetByPatientID(ctx context.Context, patientID string, page int, limit int) ([]model.Sale, int64, error) {
	oid, err := primitive.ObjectIDFromHex(patientID)
	if err != nil {
		return nil, 0, err
	}
	return u.saleRepo.FindByPatientID(ctx, oid, page, limit)
}

func (u *SaleUsecase) CheckInteractions(ctx context.Context, productIDs []string) ([]InteractionWarning, error) {
	var warnings []InteractionWarning
	var products []*model.Product

	for _, pid := range productIDs {
		oid, err := primitive.ObjectIDFromHex(pid)
		if err != nil {
			continue
		}
		product, err := u.productRepo.FindByID(ctx, oid)
		if err != nil {
			continue
		}
		products = append(products, product)
	}

	for i := 0; i < len(products); i++ {
		for j := i + 1; j < len(products); j++ {
			if hasInteraction(products[i], products[j]) {
				warnings = append(warnings, InteractionWarning{
					Drug1:   products[i].TradeName,
					Drug2:   products[j].TradeName,
					Message: products[i].TradeName + " มีปฏิกิริยากับ " + products[j].TradeName,
				})
			}
		}
	}

	return warnings, nil
}

func (u *SaleUsecase) CheckAllergies(ctx context.Context, patientID string, productIDs []string) ([]AllergyWarning, error) {
	var warnings []AllergyWarning

	pid, err := primitive.ObjectIDFromHex(patientID)
	if err != nil {
		return warnings, errors.New("invalid patient ID")
	}

	patient, err := u.patientRepo.FindByID(ctx, pid)
	if err != nil {
		return warnings, errors.New("patient not found")
	}

	if len(patient.Allergies) == 0 {
		return warnings, nil
	}

	for _, productIDStr := range productIDs {
		oid, err := primitive.ObjectIDFromHex(productIDStr)
		if err != nil {
			continue
		}
		product, err := u.productRepo.FindByID(ctx, oid)
		if err != nil {
			continue
		}

		for _, allergy := range patient.Allergies {
			allergyLower := strings.ToLower(allergy.DrugName)
			if strings.Contains(strings.ToLower(product.TradeName), allergyLower) ||
				strings.Contains(strings.ToLower(product.GenericName), allergyLower) {
				warnings = append(warnings, AllergyWarning{
					DrugName: product.TradeName,
					Reaction: allergy.Reaction,
					Severity: allergy.Severity,
				})
			}
		}
	}

	return warnings, nil
}

func hasInteraction(a *model.Product, b *model.Product) bool {
	aNameLower := strings.ToLower(a.TradeName)
	aGenericLower := strings.ToLower(a.GenericName)
	bNameLower := strings.ToLower(b.TradeName)
	bGenericLower := strings.ToLower(b.GenericName)

	for _, interaction := range a.Interactions {
		interLower := strings.ToLower(interaction)
		if strings.Contains(bNameLower, interLower) || strings.Contains(bGenericLower, interLower) {
			return true
		}
	}
	for _, interaction := range b.Interactions {
		interLower := strings.ToLower(interaction)
		if strings.Contains(aNameLower, interLower) || strings.Contains(aGenericLower, interLower) {
			return true
		}
	}
	return false
}
