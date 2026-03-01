package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Resource struct {
	PharmDb     *mongo.Database
	RdDB        *redis.Client
	MongoClient *mongo.Client
}

// Close use this method to close database connection
func (r *Resource) Close() {
	logrus.Warning("Closing all db connections")
	if r.MongoClient != nil {
		if err := r.MongoClient.Disconnect(context.Background()); err != nil {
			logrus.Errorf("Error disconnecting MongoDB: %v", err)
		}
	}
	if r.RdDB != nil {
		if err := r.RdDB.Close(); err != nil {
			logrus.Errorf("Error closing Redis: %v", err)
		}
	}
}

func InitResource() (*Resource, error) {
	// Mongo client
	host := os.Getenv("MONGO_HOST")
	pharmDbName := os.Getenv("MONGO_PHARM_DB_NAME")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(host))
	if err != nil {
		return nil, err
	}

	if err := mongoClient.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}
	log.Println("MongoDB connected successfully")

	// Redis client (optional — only used for rate limiting)
	var rdb *redis.Client
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost != "" {
		redisOp, err := redis.ParseURL(redisHost)
		if err != nil {
			log.Printf("Warning: invalid REDIS_HOST, skipping Redis: %v", err)
		} else {
			rdb = redis.NewClient(redisOp)
			_, err = rdb.Ping(context.Background()).Result()
			if err != nil {
				log.Printf("Warning: Redis not available, rate limiting disabled: %v", err)
				rdb = nil
			}
		}
	}

	return &Resource{
		PharmDb:     mongoClient.Database(pharmDbName),
		RdDB:        rdb,
		MongoClient: mongoClient,
	}, nil
}
