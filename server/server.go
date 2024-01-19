package servers

import (
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"pok92deng/config"
)

type IServer interface {
	Start(ctx context.Context) error
}

type server struct {
	app           *gin.Engine
	mongoDatabase *mongo.Database
	cfg           config.IConfig
}

func NewServer(cfg config.IConfig, mongoDatabase *mongo.Database) IServer {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(gin.Logger())

	return &server{
		app:           r,
		cfg:           cfg,
		mongoDatabase: mongoDatabase,
	}
}

func (s *server) Start(ctx context.Context) error {
	middlewares := InitMiddlewares(s)
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	config.AddAllowHeaders("Authorization", "access-control-allow-origin")
	s.app.Use(cors.New(config))
	api := s.app.Group("/api")
	modules := InitModule(api, s, middlewares)
	modules.UsersModule()
	modules.ProductsModule()
	s.app.Static("/uploads/beers", "./uploads/beers")
	mongoClient := s.mongoDatabase.Client()

	if err := mongoClient.Ping(ctx, nil); err != nil {
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
