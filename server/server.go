package servers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"pok92deng/config"
)

type IServer interface {
	Start(ctx context.Context) error
}

type server struct {
	app                *gin.Engine
	mongoClient        *mongo.Client
	usersCollection    *mongo.Collection
	productsCollection *mongo.Collection
	signsinCollection  *mongo.Collection
	cfg                config.IConfig
}

func NewServer(cfg config.IConfig, mongoClient *mongo.Client, usersCollection *mongo.Collection, productsCollection *mongo.Collection, signsinCollection *mongo.Collection) IServer {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Logger())

	return &server{
		app:                r,
		cfg:                cfg,
		mongoClient:        mongoClient,
		usersCollection:    usersCollection,
		productsCollection: productsCollection,
		signsinCollection:  signsinCollection,
	}
}

func (s *server) Start(ctx context.Context) error {
	middlewares := InitMiddlewares(s)
	api := s.app.Group("/api")
	modules := InitModule(api, s, middlewares)
	modules.UsersModule()

	if err := s.mongoClient.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	host := s.cfg.App().Host()
	port := s.cfg.App().Port()
	serverAddress := fmt.Sprintf("%s:%s", host, port)

	log.Printf("Server has been started on %s ⚡", serverAddress)

	if err := s.app.Run(serverAddress); err != nil {
		log.Fatalf("Failed to run the server: %v", err)
	}
	return nil
}
