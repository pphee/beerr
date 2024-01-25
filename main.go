package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"pok92deng/config"
	"pok92deng/module/server"
	databases "pok92deng/pkg/database"
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

	mongoDatabase, err := databases.ConnectMongoDB(cfg.Db())
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	mongoClient := mongoDatabase.Client()

	defer func() {
		err := mongoClient.Disconnect(context.Background())
		if err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	ctx := context.Background()

	server := servers.NewServer(cfg, mongoDatabase)
	if err := server.Start(ctx); err != nil { // Pass the created context
		log.Fatalf("Failed to start server: %v", err)
	}
}
