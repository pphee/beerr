package usersPatterns

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"pok92deng/users"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type IInsertUser interface {
	Customer() (IInsertUser, error)
	Admin() (IInsertUser, error)
	Result() (*users.UserPassport, error)
}

type userReq struct {
	Id  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	req *users.UserRegisterReq
	db  *mongo.Collection
}

type customer struct {
	*userReq
}

type admin struct {
	*userReq
}

func InsertUser(collection *mongo.Collection, req *users.UserRegisterReq, isAdmin bool) IInsertUser {
	if isAdmin {
		return newAdmin(collection, req)
	}

	return newCustomer(collection, req)
}

func newCustomer(collection *mongo.Collection, req *users.UserRegisterReq) IInsertUser {
	return &customer{
		userReq: &userReq{
			req: req,
			db:  collection,
		},
	}
}

func newAdmin(collection *mongo.Collection, req *users.UserRegisterReq) IInsertUser {
	return &admin{
		userReq: &userReq{
			req: req,
			db:  collection,
		},
	}
}

func (f *userReq) Customer() (IInsertUser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userDoc := bson.M{
		"email":    f.req.Email,
		"password": f.req.Password,
		"username": f.req.Username,
		"role_id":  1,
	}

	result, err := f.db.InsertOne(ctx, userDoc)
	if err != nil {
		return nil, fmt.Errorf("insert user failed: %v", err)
	}

	f.Id = result.InsertedID.(primitive.ObjectID)
	return f, nil
}

func (f *userReq) Admin() (IInsertUser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userDoc := bson.M{
		"email":    f.req.Email,
		"password": f.req.Password,
		"username": f.req.Username,
		"role_id":  2,
	}

	result, err := f.db.InsertOne(ctx, userDoc)
	if err != nil {
		return nil, fmt.Errorf("insert user failed: %v", err)
	}

	f.Id = result.InsertedID.(primitive.ObjectID)
	return f, nil

}

func (f *userReq) Result() (*users.UserPassport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": f.Id}
	var user users.User
	err := f.db.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		fmt.Println("Error in FindOne:", err)
		return nil, fmt.Errorf("get user failed: %v", err)
	}

	return &users.UserPassport{
		User: &user,
	}, nil
}
