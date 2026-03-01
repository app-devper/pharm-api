package api

import (
	"net/http"
	"pharmacy-pos/api/app/core/errs"
	"pharmacy-pos/api/middlewares"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type DashboardHandler struct {
	db *mongo.Database
}

func NewDashboardHandler(db *mongo.Database) *DashboardHandler {
	return &DashboardHandler{db: db}
}

func (h *DashboardHandler) GetStats(ctx *gin.Context) {
	clientID := ctx.GetString(middlewares.ClientId)
	c := ctx.Request.Context()

	// Today's sales
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	salesFilter := bson.M{
		"clientId":    clientID,
		"createdDate": bson.M{"$gte": startOfDay, "$lt": endOfDay},
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: salesFilter}},
		{{Key: "$group", Value: bson.M{
			"_id":        nil,
			"totalSales": bson.M{"$sum": "$total"},
			"count":      bson.M{"$sum": 1},
		}}},
	}

	var salesResult []bson.M
	cursor, err := h.db.Collection("sales").Aggregate(c, pipeline)
	if err == nil {
		cursor.All(c, &salesResult)
		cursor.Close(c)
	}

	todaySales := 0.0
	todayCount := int64(0)
	if len(salesResult) > 0 {
		if v, ok := salesResult[0]["totalSales"].(float64); ok {
			todaySales = v
		}
		if v, ok := salesResult[0]["count"].(int32); ok {
			todayCount = int64(v)
		} else if v, ok := salesResult[0]["count"].(int64); ok {
			todayCount = v
		}
	}

	// Product count
	productCount, _ := h.db.Collection("products").CountDocuments(c, bson.M{"clientId": clientID, "status": "ACTIVE"})

	// Patient count
	patientCount, _ := h.db.Collection("patients").CountDocuments(c, bson.M{"clientId": clientID})

	// Expiring batches (within 6 months)
	sixMonths := now.AddDate(0, 6, 0)
	expiringCount, _ := h.db.Collection("batches").CountDocuments(c, bson.M{
		"clientId":   clientID,
		"quantity":   bson.M{"$gt": 0},
		"expiryDate": bson.M{"$gt": now, "$lte": sixMonths},
	})

	ctx.JSON(http.StatusOK, gin.H{
		"todaySales":     todaySales,
		"todaySaleCount": todayCount,
		"productCount":   productCount,
		"patientCount":   patientCount,
		"expiringCount":  expiringCount,
	})
}

func (h *DashboardHandler) GetExpiringBatches(ctx *gin.Context) {
	clientID := ctx.GetString(middlewares.ClientId)
	c := ctx.Request.Context()

	now := time.Now()
	sixMonths := now.AddDate(0, 6, 0)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"clientId":   clientID,
			"quantity":   bson.M{"$gt": 0},
			"expiryDate": bson.M{"$gt": now, "$lte": sixMonths},
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "products",
			"localField":   "productId",
			"foreignField": "_id",
			"as":           "product",
		}}},
		{{Key: "$unwind", Value: bson.M{"path": "$product", "preserveNullAndEmptyArrays": true}}},
		{{Key: "$project", Value: bson.M{
			"_id":         1,
			"productId":   1,
			"lotNumber":   1,
			"expiryDate":  1,
			"quantity":    1,
			"productName": "$product.tradeName",
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "expiryDate", Value: 1}}}},
		{{Key: "$limit", Value: 20}},
	}

	cursor, err := h.db.Collection("batches").Aggregate(c, pipeline)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	defer cursor.Close(c)

	results := make([]bson.M, 0)
	if err := cursor.All(c, &results); err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, results)
}

func (h *DashboardHandler) GetLowStock(ctx *gin.Context) {
	clientID := ctx.GetString(middlewares.ClientId)
	c := ctx.Request.Context()

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"clientId": clientID, "quantity": bson.M{"$gt": 0}}}},
		{{Key: "$group", Value: bson.M{
			"_id":        "$productId",
			"totalStock": bson.M{"$sum": "$quantity"},
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "products",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "product",
		}}},
		{{Key: "$unwind", Value: "$product"}},
		{{Key: "$match", Value: bson.M{
			"$expr": bson.M{"$lte": bson.A{"$totalStock", "$product.minStock"}},
		}}},
		{{Key: "$project", Value: bson.M{
			"productId":   "$_id",
			"productName": "$product.tradeName",
			"totalStock":  1,
			"minStock":    "$product.minStock",
		}}},
		{{Key: "$limit", Value: 20}},
	}

	cursor, err := h.db.Collection("batches").Aggregate(c, pipeline)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	defer cursor.Close(c)

	results := make([]bson.M, 0)
	if err := cursor.All(c, &results); err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, results)
}

func (h *DashboardHandler) GetSalesSummary(ctx *gin.Context) {
	clientID := ctx.GetString(middlewares.ClientId)
	c := ctx.Request.Context()

	now := time.Now()
	sevenDaysAgo := time.Date(now.Year(), now.Month(), now.Day()-6, 0, 0, 0, 0, now.Location())

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"clientId":    clientID,
			"createdDate": bson.M{"$gte": sevenDaysAgo},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$createdDate"},
			},
			"total": bson.M{"$sum": "$total"},
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}

	cursor, err := h.db.Collection("sales").Aggregate(c, pipeline)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	defer cursor.Close(c)

	var summaryResults []bson.M
	if err := cursor.All(c, &summaryResults); err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, summaryResults)
}

func (h *DashboardHandler) GetMonthlySummary(ctx *gin.Context) {
	clientID := ctx.GetString(middlewares.ClientId)
	c := ctx.Request.Context()

	now := time.Now()
	twelveMonthsAgo := time.Date(now.Year()-1, now.Month(), 1, 0, 0, 0, 0, now.Location())

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"clientId":    clientID,
			"createdDate": bson.M{"$gte": twelveMonthsAgo},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{"format": "%Y-%m", "date": "$createdDate"},
			},
			"total": bson.M{"$sum": "$total"},
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}

	cursor, err := h.db.Collection("sales").Aggregate(c, pipeline)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	defer cursor.Close(c)

	monthlyResults := make([]bson.M, 0)
	if err := cursor.All(c, &monthlyResults); err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, monthlyResults)
}

func (h *DashboardHandler) GetGrossMargin(ctx *gin.Context) {
	clientID := ctx.GetString(middlewares.ClientId)
	c := ctx.Request.Context()

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"clientId":    clientID,
			"createdDate": bson.M{"$gte": startOfMonth},
		}}},
		{{Key: "$unwind", Value: "$items"}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "products",
			"localField":   "items.productId",
			"foreignField": "_id",
			"as":           "product",
		}}},
		{{Key: "$unwind", Value: bson.M{"path": "$product", "preserveNullAndEmptyArrays": true}}},
		{{Key: "$group", Value: bson.M{
			"_id":       nil,
			"revenue":   bson.M{"$sum": bson.M{"$multiply": bson.A{"$items.unitPrice", "$items.quantity"}}},
			"cost":      bson.M{"$sum": bson.M{"$multiply": bson.A{"$product.costPrice", "$items.quantity"}}},
			"saleCount": bson.M{"$sum": 1},
		}}},
	}

	cursor, err := h.db.Collection("sales").Aggregate(c, pipeline)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	defer cursor.Close(c)

	var marginResults []bson.M
	cursor.All(c, &marginResults)

	if len(marginResults) == 0 {
		ctx.JSON(http.StatusOK, gin.H{"revenue": 0, "cost": 0, "grossMargin": 0, "marginPercent": 0})
		return
	}

	revenue := 0.0
	cost := 0.0
	if v, ok := marginResults[0]["revenue"].(float64); ok {
		revenue = v
	}
	if v, ok := marginResults[0]["cost"].(float64); ok {
		cost = v
	}
	margin := revenue - cost
	pct := 0.0
	if revenue > 0 {
		pct = (margin / revenue) * 100
	}

	ctx.JSON(http.StatusOK, gin.H{
		"revenue":       revenue,
		"cost":          cost,
		"grossMargin":   margin,
		"marginPercent": pct,
	})
}

func (h *DashboardHandler) GetABCAnalysis(ctx *gin.Context) {
	clientID := ctx.GetString(middlewares.ClientId)
	c := ctx.Request.Context()

	sixMonthsAgo := time.Now().AddDate(0, -6, 0)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"clientId":    clientID,
			"createdDate": bson.M{"$gte": sixMonthsAgo},
		}}},
		{{Key: "$unwind", Value: "$items"}},
		{{Key: "$group", Value: bson.M{
			"_id":        "$items.productId",
			"totalQty":   bson.M{"$sum": "$items.quantity"},
			"totalSales": bson.M{"$sum": bson.M{"$multiply": bson.A{"$items.unitPrice", "$items.quantity"}}},
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "products",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "product",
		}}},
		{{Key: "$unwind", Value: "$product"}},
		{{Key: "$project", Value: bson.M{
			"productId":   "$_id",
			"productName": "$product.tradeName",
			"totalQty":    1,
			"totalSales":  1,
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "totalSales", Value: -1}}}},
	}

	cursor, err := h.db.Collection("sales").Aggregate(c, pipeline)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	defer cursor.Close(c)

	var items []bson.M
	cursor.All(c, &items)

	grandTotal := 0.0
	for _, item := range items {
		if v, ok := item["totalSales"].(float64); ok {
			grandTotal += v
		}
	}

	cumulative := 0.0
	abcResult := make([]bson.M, 0)
	for _, item := range items {
		sales := 0.0
		if v, ok := item["totalSales"].(float64); ok {
			sales = v
		}
		cumulative += sales
		cpct := 0.0
		if grandTotal > 0 {
			cpct = (cumulative / grandTotal) * 100
		}
		cls := "C"
		if cpct <= 80 {
			cls = "A"
		} else if cpct <= 95 {
			cls = "B"
		}
		item["class"] = cls
		item["cumulativePercent"] = cpct
		abcResult = append(abcResult, item)
	}

	ctx.JSON(http.StatusOK, abcResult)
}

func (h *DashboardHandler) GetDeadStock(ctx *gin.Context) {
	clientID := ctx.GetString(middlewares.ClientId)
	c := ctx.Request.Context()

	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)

	soldPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"clientId":    clientID,
			"createdDate": bson.M{"$gte": ninetyDaysAgo},
		}}},
		{{Key: "$unwind", Value: "$items"}},
		{{Key: "$group", Value: bson.M{
			"_id": "$items.productId",
		}}},
	}

	soldCursor, err := h.db.Collection("sales").Aggregate(c, soldPipeline)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	var soldResults []bson.M
	soldCursor.All(c, &soldResults)
	soldCursor.Close(c)

	soldIDs := make([]interface{}, 0)
	for _, r := range soldResults {
		soldIDs = append(soldIDs, r["_id"])
	}

	filter := bson.M{
		"clientId": clientID,
		"status":   "ACTIVE",
	}
	if len(soldIDs) > 0 {
		filter["_id"] = bson.M{"$nin": soldIDs}
	}

	productCursor, err := h.db.Collection("products").Find(c, filter)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	defer productCursor.Close(c)

	var products []bson.M
	productCursor.All(c, &products)

	deadResult := make([]bson.M, 0)
	for _, p := range products {
		pid := p["_id"]
		stockPipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.M{"productId": pid, "clientId": clientID, "quantity": bson.M{"$gt": 0}}}},
			{{Key: "$group", Value: bson.M{"_id": nil, "totalStock": bson.M{"$sum": "$quantity"}}}},
		}
		stockCursor, _ := h.db.Collection("batches").Aggregate(c, stockPipeline)
		var stockRes []bson.M
		stockCursor.All(c, &stockRes)
		stockCursor.Close(c)

		stock := 0
		if len(stockRes) > 0 {
			if v, ok := stockRes[0]["totalStock"].(int32); ok {
				stock = int(v)
			} else if v, ok := stockRes[0]["totalStock"].(int64); ok {
				stock = int(v)
			}
		}
		if stock > 0 {
			deadResult = append(deadResult, bson.M{
				"productId":   pid,
				"productName": p["tradeName"],
				"totalStock":  stock,
				"costPrice":   p["costPrice"],
			})
		}
	}

	ctx.JSON(http.StatusOK, deadResult)
}

func (h *DashboardHandler) GetRefillReminders(ctx *gin.Context) {
	clientID := ctx.GetString(middlewares.ClientId)
	c := ctx.Request.Context()

	sixtyDaysAgo := time.Now().AddDate(0, 0, -60)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"clientId":    clientID,
			"createdDate": bson.M{"$gte": sixtyDaysAgo},
			"patientId":   bson.M{"$exists": true, "$ne": nil},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "createdDate", Value: -1}}}},
		{{Key: "$group", Value: bson.M{
			"_id":          "$patientId",
			"lastDispense": bson.M{"$first": "$createdDate"},
			"items":        bson.M{"$first": "$items"},
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "patients",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "patient",
		}}},
		{{Key: "$unwind", Value: "$patient"}},
		{{Key: "$project", Value: bson.M{
			"patientId":    "$_id",
			"patientName":  bson.M{"$concat": bson.A{"$patient.firstName", " ", "$patient.lastName"}},
			"phone":        "$patient.phone",
			"lastDispense": 1,
			"itemCount":    bson.M{"$size": "$items"},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "lastDispense", Value: 1}}}},
		{{Key: "$limit", Value: 20}},
	}

	cursor, err := h.db.Collection("sales").Aggregate(c, pipeline)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	defer cursor.Close(c)

	refillResults := make([]bson.M, 0)
	cursor.All(c, &refillResults)

	now := time.Now()
	for i, r := range refillResults {
		if lastDispense, ok := r["lastDispense"].(time.Time); ok {
			daysSince := int(now.Sub(lastDispense).Hours() / 24)
			refillDate := lastDispense.AddDate(0, 0, 30)
			refillResults[i]["daysSinceLastDispense"] = daysSince
			refillResults[i]["estimatedRefillDate"] = refillDate
			refillResults[i]["isOverdue"] = now.After(refillDate)
		}
	}

	ctx.JSON(http.StatusOK, refillResults)
}
