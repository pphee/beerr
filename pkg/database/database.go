package databases

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"pok92deng/config"
)

func ConnectMongoDB(cfg config.IDbConfig) (*mongo.Client, *mongo.Collection, *mongo.Collection, *mongo.Collection, error) {
	clientOptions := options.Client().ApplyURI(cfg.Url())
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(cfg.Name())
	userscollection := db.Collection(cfg.UsersCollection())
	signinscollection := db.Collection(cfg.SigninsCollection())
	productscollection := db.Collection(cfg.ProductsCollection())
	return client, userscollection, productscollection, signinscollection, nil
}
