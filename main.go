package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"pok92deng/config"
	"pok92deng/module/server"
	auth "pok92deng/pkg"
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
	godotenv.Load("/Users/p/Goland/pok92deng/.env")
	cfg := config.LoadConfig("/Users/p/Goland/pok92deng/.env")
	fmt.Println(cfg.Db())

	ctx := context.Background()

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

	zitadel, err := auth.InitZitadelAuthorization(ctx, cfg.Zitadel())
	if err != nil {
		log.Fatalf("Failed to initialize ZITADEL: %v", err)
	}

	connectZitadel, err := databases.ConnectZitadel(ctx, cfg.Zitadel())

	server := servers.NewServer(cfg, mongoDatabase, zitadel, connectZitadel)
	if err := server.Start(ctx); err != nil { // Pass the created context
		log.Fatalf("Failed to start server: %v", err)
	}
}
