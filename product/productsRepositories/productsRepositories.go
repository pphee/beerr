package repository

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"pok92deng/config"
	model "pok92deng/product"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type BeerRepository interface {
	InsertBeer(ctx context.Context, beer model.Beer) (primitive.ObjectID, error)
	FindBeer(ctx context.Context, id primitive.ObjectID) (model.Beer, error)
	UpdateBeer(ctx context.Context, id primitive.ObjectID, beer model.Beer) error
	DeleteBeer(ctx context.Context, id primitive.ObjectID) error
	FilterAndPaginateBeers(ctx context.Context, name string, page, limit int64) ([]*model.Beer, int64, error)
	BeerNameExists(ctx context.Context, name string) (bool, error)
}

type productsRepository struct {
	db  *mongo.Database
	cfg config.IConfig
}

func NewproductsRepository(cfg config.IConfig, mongoDatabase *mongo.Database) BeerRepository {
	if mongoDatabase == nil {
		return nil
	}
	return &productsRepository{
		cfg: cfg,
		db:  mongoDatabase,
	}
}

func (r *productsRepository) BeerNameExists(ctx context.Context, name string) (bool, error) {
	var beer model.Beer
	if name == "" {
		return false, errors.New("name cannot be empty")
	}
	filter := bson.M{"name": name}
	err := r.db.Collection(r.cfg.Db().ProductsCollection()).FindOne(ctx, filter).Decode(&beer)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *productsRepository) InsertBeer(ctx context.Context, beer model.Beer) (primitive.ObjectID, error) {
	exists, err := r.BeerNameExists(ctx, beer.Name)
	if err != nil {
		return primitive.NilObjectID, err
	}
	if exists {
		return primitive.NilObjectID, errors.New("a beer with this name already exists")
	}

	result, err := r.db.Collection(r.cfg.Db().ProductsCollection()).InsertOne(ctx, beer)

	if err != nil {
		return primitive.NilObjectID, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func (r *productsRepository) FindBeer(ctx context.Context, id primitive.ObjectID) (model.Beer, error) {
	var beer model.Beer
	if err := r.db.Collection(r.cfg.Db().ProductsCollection()).FindOne(ctx, bson.M{"_id": id}).Decode(&beer); err != nil {
		return model.Beer{}, err
	}
	return beer, nil
}

func (r *productsRepository) UpdateBeer(ctx context.Context, id primitive.ObjectID, beer model.Beer) error {
	if beer.Image == nil || beer.ImagePath == "" {
		existingBeer, err := r.FindBeer(ctx, id)
		if err != nil {
			return err
		}

		if beer.Image == nil {
			beer.Image = existingBeer.Image
		}

		if beer.ImagePath == "" {
			beer.ImagePath = existingBeer.ImagePath
		}
	}

	update := bson.M{"$set": beer}
	result, err := r.db.Collection(r.cfg.Db().ProductsCollection()).UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("no beer found with specified ID")
	}
	return nil
}

func (r *productsRepository) DeleteBeer(ctx context.Context, id primitive.ObjectID) error {

	result, err := r.db.Collection(r.cfg.Db().ProductsCollection()).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("no beer found with specified ID")
	}

	return nil
}

func (r *productsRepository) FilterAndPaginateBeers(ctx context.Context, name string, page, limit int64) ([]*model.Beer, int64, error) {
	var beers []*model.Beer
	skip := (page - 1) * limit

	filter := bson.M{}
	if name != "" {
		filter["name"] = bson.M{"$regex": name, "$options": "i"}
	}

	total, err := r.db.Collection(r.cfg.Db().ProductsCollection()).CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOptions := options.Find()
	findOptions.SetLimit(limit)
	findOptions.SetSkip(skip)

	cursor, err := r.db.Collection(r.cfg.Db().ProductsCollection()).Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching beers: %w", err)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			fmt.Println("Error closing cursor:", err)
		}
	}(cursor, ctx)

	for cursor.Next(ctx) {
		var beer model.Beer
		if err := cursor.Decode(&beer); err != nil {
			return nil, 0, err
		}
		beers = append(beers, &beer)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	return beers, total, nil
}
