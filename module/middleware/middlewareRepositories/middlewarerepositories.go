package middlewaresRepositories

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"pok92deng/config"
	"pok92deng/module/middleware"
)

type IMiddlewaresRepository interface {
	FindAccessToken(userId, accessToken string) bool
	FindRole(ctx context.Context, userRoleId string) ([]*middlewares.Roles, error)
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
	count, err := r.db.Collection(r.cfg.Db().SignInsCollection()).CountDocuments(context.Background(), filter)
	if err != nil {
		fmt.Println("Error counting documents:", err)
		return false
	}

	return count > 0
}

func (r *middlewaresRepository) FindRole(ctx context.Context, userRoleId string) ([]*middlewares.Roles, error) {
	var filter bson.D
	if userRoleId != "" {
		filter = bson.D{{"roleId", userRoleId}}
	}

	cursor, err := r.db.Collection(r.cfg.Db().RolesCollection()).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error querying roles: %v", err)
	}
	defer cursor.Close(ctx)

	var roles []*middlewares.Roles
	if err = cursor.All(ctx, &roles); err != nil {
		return nil, fmt.Errorf("error decoding roles: %v", err)
	}

	return roles, nil
}
