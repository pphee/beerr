package usecases

import (
	"context"
	model "pok92deng/product"
	repository "pok92deng/product/productsRepositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BeerService interface {
	CreateBeer(ctx context.Context, beer model.Beer) (primitive.ObjectID, error)
	GetBeer(ctx context.Context, id primitive.ObjectID) (model.Beer, error)
	UpdateBeer(ctx context.Context, id primitive.ObjectID, beer model.Beer) error
	DeleteBeer(ctx context.Context, id primitive.ObjectID) error
	ListBeers(ctx context.Context) ([]model.Beer, error)
	FilterBeersByName(ctx context.Context, name string) ([]model.Beer, error)
	Pagination(page, limit int64) ([]*model.Beer, int64, error)
}

type beerServiceImpl struct {
	repo repository.BeerRepository
}

func NewBeerService(repo repository.BeerRepository) BeerService {
	return &beerServiceImpl{repo: repo}
}

func (s *beerServiceImpl) CreateBeer(ctx context.Context, beer model.Beer) (primitive.ObjectID, error) {

	return s.repo.InsertBeer(ctx, beer)
}

func (s *beerServiceImpl) GetBeer(ctx context.Context, id primitive.ObjectID) (model.Beer, error) {
	return s.repo.FindBeer(ctx, id)
}

func (s *beerServiceImpl) UpdateBeer(ctx context.Context, id primitive.ObjectID, beer model.Beer) error {
	return s.repo.UpdateBeer(ctx, id, beer)
}

func (s *beerServiceImpl) DeleteBeer(ctx context.Context, id primitive.ObjectID) error {
	return s.repo.DeleteBeer(ctx, id)
}

func (s *beerServiceImpl) ListBeers(ctx context.Context) ([]model.Beer, error) {
	return s.repo.ListBeers(ctx)
}

func (s *beerServiceImpl) FilterBeersByName(ctx context.Context, name string) ([]model.Beer, error) {
	return s.repo.FilterBeersByName(ctx, name)
}

func (s *beerServiceImpl) Pagination(page, limit int64) ([]*model.Beer, int64, error) {
	return s.repo.Pagination(page, limit)
}
