package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	TripsCollection     = "trips"
	NexusDriveGoCollection = "NexusDriveGo"
)

// MongoDB connection configuration
type MongoConfig struct {
	URI      string
	Database string
}

func NewMongoDefaultConfig() *MongoConfig {
	return &MongoConfig{
		URI:      os.Getenv("MONGODB_URI"),
		Database: "NexusDriveGo",
	}
}

func NewMongoClient(ctx context.Context, cfg *MongoConfig) (*mongo.Client, error) {
	if cfg.URI == "" {
		return nil, fmt.Errorf("mongodb URI is required")
	}
	if cfg.Database == "" {
		return nil, fmt.Errorf("mongodb database is required")
	}

	connCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(connCtx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}

	err = client.Ping(connCtx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	log.Printf("Successfully connected to MongoDB at %s", cfg.URI)
	return client, nil
}

func GetDatabase(client *mongo.Client, cfg *MongoConfig) *mongo.Database {
	return client.Database(cfg.Database)
}