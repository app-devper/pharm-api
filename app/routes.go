package app

import (
	"context"
	"log"
	"os"
	"pharmacy-pos/api/app/domain/model"
	"pharmacy-pos/api/app/domain/usecase"
	"pharmacy-pos/api/app/features/api"
	"pharmacy-pos/api/app/features/repo"
	"pharmacy-pos/api/db"
	"pharmacy-pos/api/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Helper function to check if a feature is enabled
func isFeatureEnabled(db *mongo.Database, key string) bool {
	var setting model.Setting
	err := db.Collection("settings").FindOne(context.Background(), bson.M{"key": key}).Decode(&setting)
	if err != nil {
		return false // Default to disabled if setting not found
	}
	return setting.Value == "true"
}

type Routes struct{}

func (r Routes) StartGin() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	resources, err := db.InitResource()
	if err != nil {
		log.Fatalf("Failed to initialize database resources: %v", err)
	}
	defer resources.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	gin.SetMode(os.Getenv("GIN_MODE"))

	app := gin.New()
	app.SetTrustedProxies(nil)

	app.Use(gin.Logger())
	app.Use(middlewares.NewRecovery())
	app.Use(middlewares.NewCors([]string{"*"}))

	// Init repositories
	productRepo := repo.NewProductRepo(resources.PharmDb)
	patientRepo := repo.NewPatientRepo(resources.PharmDb)
	saleRepo := repo.NewSaleRepo(resources.PharmDb)
	batchRepo := repo.NewBatchRepo(resources.PharmDb)

	// Init usecases
	productUC := usecase.NewProductUsecase(productRepo)
	patientUC := usecase.NewPatientUsecase(patientRepo)
	saleUC := usecase.NewSaleUsecase(saleRepo, productRepo, patientRepo, batchRepo, resources.MongoClient)
	batchUC := usecase.NewBatchUsecase(batchRepo, productRepo)

	// Init handlers
	productHandler := api.NewProductHandler(productUC)
	patientHandler := api.NewPatientHandler(patientUC)
	saleHandler := api.NewSaleHandler(saleUC)
	batchHandler := api.NewBatchHandler(batchUC)
	dashboardHandler := api.NewDashboardHandler(resources.PharmDb)
	reportHandler := api.NewReportHandler(resources.PharmDb)
	settingHandler := api.NewSettingHandler(resources.PharmDb)
	receiveHandler := api.NewReceiveHandler(resources.PharmDb, batchUC)

	// API routes
	v1 := app.Group("/api/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
				"db":      resources.PharmDb.Name(),
			})
		})
	}

	// Protected routes
	auth := v1.Group("")
	auth.Use(middlewares.RequireAuthenticated())
	{
		// Products
		products := auth.Group("/products")
		{
			products.POST("", productHandler.Create)
			products.GET("", productHandler.GetAll)
			products.GET("/:id", productHandler.GetByID)
			products.GET("/barcode/:barcode", productHandler.GetByBarcode)
			products.PUT("/:id", productHandler.Update)
			products.DELETE("/:id", productHandler.Delete)
		}

		// Patients (feature toggle controlled)
		if isFeatureEnabled(resources.PharmDb, "patient_feature_enabled") {
			patients := auth.Group("/patients")
			{
				patients.POST("", patientHandler.Create)
				patients.GET("", patientHandler.GetAll)
				patients.GET("/:id", patientHandler.GetByID)
				patients.PUT("/:id", patientHandler.Update)
				patients.DELETE("/:id", patientHandler.Delete)
			}
			// Patient history
			auth.GET("/patients/:id/history", saleHandler.GetPatientHistory)
		}

		// Sales
		sales := auth.Group("/sales")
		{
			sales.POST("", saleHandler.Create)
			sales.GET("", saleHandler.GetAll)
			sales.GET("/:id", saleHandler.GetByID)
			sales.POST("/check-interactions", saleHandler.CheckInteractions)
			sales.POST("/check-allergies", saleHandler.CheckAllergies)
		}

		// Batches / Inventory
		batches := auth.Group("/batches")
		{
			batches.GET("", batchHandler.GetAll)
			batches.POST("", batchHandler.Create)
			batches.GET("/product/:productId", batchHandler.GetByProductID)
			batches.GET("/expiring", batchHandler.GetExpiringBatches)
			batches.GET("/low-stock", batchHandler.GetLowStock)
			batches.PUT("/:id", batchHandler.Update)
			batches.DELETE("/:id", batchHandler.Delete)
		}

		// Dashboard
		dashboard := auth.Group("/dashboard")
		{
			dashboard.GET("/stats", dashboardHandler.GetStats)
			dashboard.GET("/sales-summary", dashboardHandler.GetSalesSummary)
			dashboard.GET("/monthly-summary", dashboardHandler.GetMonthlySummary)
			dashboard.GET("/gross-margin", dashboardHandler.GetGrossMargin)
			dashboard.GET("/abc-analysis", dashboardHandler.GetABCAnalysis)
			dashboard.GET("/dead-stock", dashboardHandler.GetDeadStock)
			dashboard.GET("/refill-reminders", dashboardHandler.GetRefillReminders)
			dashboard.GET("/expiring", dashboardHandler.GetExpiringBatches)
			dashboard.GET("/low-stock", dashboardHandler.GetLowStock)
		}

		// Reports (Admin + Pharmacist only)
		reports := auth.Group("/reports")
		reports.Use(middlewares.RequireAuthorization("SUPER", "ADMIN"))
		{
			reports.GET("/ky9", reportHandler.GetKY9)   // บัญชีซื้อยา (receive)
			reports.GET("/ky10", reportHandler.GetKY10) // ขายยาอันตราย
			reports.GET("/ky11", reportHandler.GetKY11) // ขายยาควบคุมพิเศษ
			reports.GET("/ky12", reportHandler.GetKY12) // วัตถุออกฤทธิ์ฯ
			reports.GET("/ky13", reportHandler.GetKY13) // ยาเสพติดให้โทษ ประเภท 3
		}

		// Receives (Goods Receipts) - Admin only
		receives := auth.Group("/receives")
		receives.Use(middlewares.RequireAuthorization("SUPER", "ADMIN"))
		{
			receives.POST("", receiveHandler.Create)
			receives.GET("", receiveHandler.GetAll)
			receives.GET("/:id", receiveHandler.GetByID)
		}

		// Settings (Admin only)
		settings := auth.Group("/settings")
		settings.Use(middlewares.RequireAuthorization("SUPER", "ADMIN"))
		{
			settings.GET("", settingHandler.GetAll)
			settings.GET("/:key", settingHandler.GetByKey)
			settings.PUT("/:key", settingHandler.Upsert)
		}
	}

	app.NoRoute(middlewares.NoRoute())

	app.Run(":" + port)
}
