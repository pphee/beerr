package middlewaresRepositories

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	middlewares "pok92deng/middleware"
)

type IMiddlewaresRepository interface {
	FindAccessToken(userId, accessToken string) bool
	FindRole() ([]*middlewares.Role, error)
}

type middlewaresRepository struct {
	db *mongo.Collection
}

func MiddlewaresRepository(db *mongo.Collection) IMiddlewaresRepository {
	return &middlewaresRepository{
		db: db,
	}
}

func (r *middlewaresRepository) FindAccessToken(userId, accessToken string) bool {
	objId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		fmt.Println("Error converting userId to ObjectId:", err)
		return false
	}

	filter := bson.M{"user_id": objId, "access_token": accessToken}

	count, err := r.db.CountDocuments(context.Background(), filter)
	if err != nil {
		fmt.Println("Error counting documents:", err)
		return false
	}

	return count > 0
}

func (r *middlewaresRepository) FindRole() ([]*middlewares.Role, error) {
	cursor, err := r.db.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, fmt.Errorf("roles are empty")
	}

	var roles []*middlewares.Role
	if err = cursor.All(context.Background(), &roles); err != nil {
		return nil, err
	}
	return roles, nil
}
