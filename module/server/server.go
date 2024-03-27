package servers

import (
	"context"
	"flag"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/http/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"pok92deng/config"
)

type IServer interface {
	Start(ctx context.Context) error
}

type server struct {
	app            *gin.Engine
	mongoDatabase  *mongo.Database
	cfg            config.IConfig
	zitadel        *authorization.Authorizer[*oauth.IntrospectionContext]
	connectZitadel *client.Client
}

func NewServer(cfg config.IConfig, mongoDatabase *mongo.Database, zitadel *authorization.Authorizer[*oauth.IntrospectionContext], connectZitadel *client.Client) IServer {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(gin.Logger())

	return &server{
		app:            r,
		cfg:            cfg,
		mongoDatabase:  mongoDatabase,
		zitadel:        zitadel,
		connectZitadel: connectZitadel,
	}
}

func (s *server) Start(ctx context.Context) error {

	flag.Parse()
	middlewares := InitMiddlewares(s)
	mw := middleware.New(s.zitadel)

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	config.AddAllowHeaders("Authorization", "access-control-allow-origin")
	flag.Parse()

	s.app.Use(cors.New(config))
	api := s.app.Group("/api")

	modules := InitModule(api, s, middlewares, mw)
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
