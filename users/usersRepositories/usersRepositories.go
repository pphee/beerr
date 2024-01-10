package usersRepositories

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"pok92deng/config"
	"pok92deng/users"
	"pok92deng/users/usersPatterns"
	"time"
)

type UserRepository interface {
	InsertUser(req *users.UserRegisterReq, isAdmin bool) (*users.UserPassport, error)
	FindOneUserByEmail(email string) (*users.UserCredentialCheck, error)
	InsertOauth(req *users.UserPassport) error
	FindOneOauth(refreshToken string) (*users.Oauth, error)
	GetProfile(userId string) (*users.User, error)
	UpdateOauth(req *users.UserToken) error
}

type usersRepository struct {
	db  *mongo.Database
	cfg config.IConfig
}

func UsersRepository(cfg config.IConfig, mongoDatabase *mongo.Database) UserRepository {
	return &usersRepository{
		cfg: cfg,
		db:  mongoDatabase,
	}
}

func (r *usersRepository) InsertUser(req *users.UserRegisterReq, isAdmin bool) (*users.UserPassport, error) {
	if err := r.checkUserExistence(req.Email, req.Username); err != nil {
		return nil, err
	}

	result := usersPatterns.InsertUser(r.db.Collection(r.cfg.Db().UsersCollection()), req, isAdmin)

	var err error
	if isAdmin {
		result, err = result.Admin()
	} else {
		result, err = result.Customer()
	}
	if err != nil {
		return nil, err
	}

	user, err := result.Result()
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *usersRepository) checkUserExistence(email, username string) error {
	if user, err := r.FindOneUserByEmail(email); err != nil && err.Error() != "user not found" {
		return fmt.Errorf("error checking for existing user by email: %v", err)
	} else if user != nil {
		return users.NewHTTPError(409, "a user with the same email already exists")
	}

	if user, err := r.FindOneUserByUsername(username); err != nil && err.Error() != "user not found" {
		return fmt.Errorf("error checking for existing user by username: %v", err)
	} else if user != nil {
		return users.NewHTTPError(409, "a user with the same username already exists")
	}

	return nil
}

func (r *usersRepository) FindOneUser(filter bson.M) (*users.UserCredentialCheck, error) {
	var user users.UserCredentialCheck

	err := r.db.Collection(r.cfg.Db().UsersCollection()).FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		if errors.Is(mongo.ErrNoDocuments, err) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (r *usersRepository) FindOneUserByUsername(username string) (*users.UserCredentialCheck, error) {
	filter := bson.M{"username": username}
	return r.FindOneUser(filter)
}

func (r *usersRepository) FindOneUserByEmail(email string) (*users.UserCredentialCheck, error) {
	filter := bson.M{"email": email}
	return r.FindOneUser(filter)
}

func (r *usersRepository) InsertOauth(req *users.UserPassport) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	oauthDocument := bson.M{
		"user_id":       req.User.Id,
		"refresh_token": req.Token.RefreshToken,
		"access_token":  req.Token.AccessToken,
		"created_at":    time.Now(),
		"updated_at":    time.Now(),
	}

	result, err := r.db.Collection(r.cfg.Db().SigninsCollection()).InsertOne(ctx, oauthDocument)
	if err != nil {
		return fmt.Errorf("insert oauth failed: %v", err)
	}

	// Assuming req.Token.Id is the field to store the generated ID
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		req.Token.Id = oid.Hex() // Convert ObjectID to hex string
	} else {
		return fmt.Errorf("failed to get inserted ID")
	}

	return nil
}

func (r *usersRepository) FindOneOauth(refreshToken string) (*users.Oauth, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Define a struct that matches the MongoDB document structure

	var mongoPassport users.MongoPassport
	filter := bson.M{"refresh_token": refreshToken}

	err := r.db.Collection(r.cfg.Db().SigninsCollection()).FindOne(ctx, filter).Decode(&mongoPassport)
	if err != nil {
		fmt.Println("Error:", err)
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("oauth not found: %v", err)
		}
		return nil, fmt.Errorf("error finding oauth: %v", err)
	}

	// Construct UserPassport from the MongoPassport data
	userPassport := users.UserPassport{
		User: &users.User{
			Id: mongoPassport.UserID,
		},
		Token: &users.UserToken{
			Id:           mongoPassport.ID.Hex(),
			AccessToken:  mongoPassport.AccessToken,
			RefreshToken: mongoPassport.RefreshToken,
		},
	}

	return &users.Oauth{
		Id:     userPassport.Token.Id,
		UserId: userPassport.User.Id.Hex(),
	}, nil
}

func (r *usersRepository) GetProfile(userId string) (*users.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var profile users.User

	objID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %v", err)
	}

	filter := bson.M{"_id": objID}

	if err := r.db.Collection(r.cfg.Db().UsersCollection()).FindOne(ctx, filter).Decode(&profile); err != nil {
		if errors.Is(mongo.ErrNoDocuments, err) {
			return nil, fmt.Errorf("get user failed: user not found")
		}
		return nil, fmt.Errorf("get user failed: %v", err)
	}

	return &profile, nil
}

func (r *usersRepository) UpdateOauth(req *users.UserToken) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	fmt.Println(req.Id)
	objID, err := primitive.ObjectIDFromHex(req.Id)

	if err != nil {
		return fmt.Errorf("invalid ID format: %v", err)
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$set": bson.M{
			"access_token":  req.AccessToken,
			"refresh_token": req.RefreshToken,
		},
	}

	result, err := r.db.Collection(r.cfg.Db().SigninsCollection()).UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("update oauth failed: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("update oauth failed: no document found with the given ID")
	}

	return nil
}