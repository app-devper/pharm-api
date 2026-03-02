package api

import (
	"net/http"
	"pharmacy-pos/api/app/core/errs"
	"pharmacy-pos/api/middlewares"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SettingHandler struct {
	db *mongo.Database
}

func NewSettingHandler(db *mongo.Database) *SettingHandler {
	return &SettingHandler{db: db}
}

func (h *SettingHandler) GetAll(ctx *gin.Context) {
	c := ctx.Request.Context()

	cursor, err := h.db.Collection("settings").Find(c, bson.M{})
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}
	defer cursor.Close(c)

	results := make([]bson.M, 0)
	cursor.All(c, &results)

	ctx.JSON(http.StatusOK, results)
}

func (h *SettingHandler) GetByKey(ctx *gin.Context) {
	key := ctx.Param("key")
	c := ctx.Request.Context()

	var result bson.M
	err := h.db.Collection("settings").FindOne(c, bson.M{"key": key}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusOK, gin.H{"key": key, "value": ""})
			return
		}
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (h *SettingHandler) Upsert(ctx *gin.Context) {
	updatedBy := ctx.GetString(middlewares.SessionId)
	key := ctx.Param("key")

	var body struct {
		Value string `json:"value"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		errs.Response(ctx, http.StatusBadRequest, errs.New(errs.ErrBadRequest, "invalid request body"))
		return
	}

	c := ctx.Request.Context()
	filter := bson.M{"key": key}
	update := bson.M{
		"$set": bson.M{
			"value":       body.Value,
			"updatedBy":   updatedBy,
			"updatedDate": time.Now(),
		},
		"$setOnInsert": bson.M{
			"key": key,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := h.db.Collection("settings").UpdateOne(c, filter, update, opts)
	if err != nil {
		errs.Response(ctx, http.StatusInternalServerError, errs.New(errs.ErrInternal, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "setting updated"})
}
