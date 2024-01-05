package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"os"
	"pok92deng/config"
	databases "pok92deng/pkg/database"
	servers "pok92deng/server"
)

func envPath() string {
	if len(os.Args) == 1 {
		return ".env"
	} else {
		return os.Args[1]
	}
}

func main() {
	cfg := config.LoadConfig(envPath())
	fmt.Println(cfg.Db())

	mongoClient, usersCollection, productsCollection, signinsCollection, err := databases.ConnectMongoDB(cfg.Db())

	if err != nil {
		panic(err)
	}
	defer func(mongoClient *mongo.Client, ctx context.Context) {
		err := mongoClient.Disconnect(ctx)
		if err != nil {

		}
	}(mongoClient, nil)
	ctx := context.Background()

	server := servers.NewServer(cfg, mongoClient, usersCollection, productsCollection, signinsCollection)
	if err := server.Start(ctx); err != nil { // Pass the created context
		log.Fatalf("Failed to start server: %v", err)
	}
}
