package middlewaresRepositories

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"pok92deng/config"
	middlewares "pok92deng/middleware"
)

type IMiddlewaresRepository interface {
	FindAccessToken(userId, accessToken string) bool
	FindRole() ([]*middlewares.Role, error)
}

type middlewaresRepository struct {
	cfg config.IConfig
	db  *mongo.Database
}

func MiddlewaresRepository(cfg config.IConfig, db *mongo.Database) IMiddlewaresRepository {
	return &middlewaresRepository{
		cfg: cfg,
		db:  db,
	}
}

func (r *middlewaresRepository) FindAccessToken(userId, accessToken string) bool {
	objId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		fmt.Println("Error converting userId to ObjectId:", err)
		return false
	}

	filter := bson.M{"user_id": objId, "access_token": accessToken}
	count, err := r.db.Collection(r.cfg.Db().SigninsCollection()).CountDocuments(context.Background(), filter)
	if err != nil {
		fmt.Println("Error counting documents:", err)
		return false
	}

	return count > 0
}

func (r *middlewaresRepository) FindRole() ([]*middlewares.Role, error) {
	cursor, err := r.db.Collection(r.cfg.Db().RolesCollection()).Find(context.Background(), bson.D{})
	if err != nil {
		return nil, fmt.Errorf("roles are empty")
	}

	var roles []*middlewares.Role
	if err = cursor.All(context.Background(), &roles); err != nil {
		return nil, err
	}
	return roles, nil
}
