package usecases

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mime/multipart"
	model "pok92deng/product"
	"testing"
)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) BeerNameExists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *mockRepository) InsertBeer(ctx context.Context, beer model.Beer) (primitive.ObjectID, error) {
	args := m.Called(ctx, beer)
	return args.Get(0).(primitive.ObjectID), args.Error(1)
}

func (m *mockRepository) FindBeer(ctx context.Context, id primitive.ObjectID) (model.Beer, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Beer), args.Error(1)
}

func (m *mockRepository) UpdateBeer(ctx context.Context, id primitive.ObjectID, beer model.Beer) error {
	args := m.Called(ctx, id, beer)
	return args.Error(0)
}

func (m *mockRepository) DeleteBeer(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockRepository) FilterAndPaginateBeers(ctx context.Context, name string, page, limit int64) ([]*model.Beer, int64, error) {
	args := m.Called(ctx, name, page, limit)
	return args.Get(0).([]*model.Beer), args.Get(1).(int64), args.Error(2)
}

func (m *mockRepository) UploadImage(ctx context.Context, image model.UploadBeerImageResponse) (primitive.ObjectID, error) {
	args := m.Called(ctx, image)
	return args.Get(0).(primitive.ObjectID), args.Error(1)
}

func (m *mockRepository) UpdateImage(ctx context.Context, id primitive.ObjectID, image model.UploadBeerImageResponse) error {
	args := m.Called(ctx, id, image)
	return args.Error(0)
}

func (m *mockRepository) GetUpdateBeer(ctx context.Context, id primitive.ObjectID, beer model.BeerUpdate) error {
	args := m.Called(ctx, id, beer)
	return args.Error(0)
}

func TestBeerService(t *testing.T) {
	fileHeader := &multipart.FileHeader{
		Filename: "test_beer_image.jpg",
		Header:   make(map[string][]string),
	}
	mockRepo := new(mockRepository)
	service := NewBeerService(mockRepo)

	testBeer := model.Beer{
		ID:        primitive.NewObjectID(),
		Name:      "Test Beer",
		Category:  "Ale",
		Detail:    "Test Detail",
		Image:     fileHeader,
		ImagePath: "test.jpg",
	}

	t.Run("CreateBeer", func(t *testing.T) {
		mockRepo.On("InsertBeer", mock.Anything, testBeer).Return(testBeer.ID, nil)
		id, err := service.CreateBeer(context.Background(), testBeer)
		assert.NoError(t, err)
		assert.Equal(t, testBeer.ID, id)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetBeer", func(t *testing.T) {
		mockRepo.On("FindBeer", mock.Anything, testBeer.ID).Return(testBeer, nil)
		result, err := service.GetBeer(context.Background(), testBeer.ID)
		assert.NoError(t, err)
		assert.Equal(t, testBeer, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("UpdateBeer", func(t *testing.T) {
		updatedBeer := testBeer
		updatedBeer.Name = "Updated Beer"
		mockRepo.On("UpdateBeer", mock.Anything, testBeer.ID, updatedBeer).Return(nil)
		err := service.UpdateBeer(context.Background(), testBeer.ID, updatedBeer)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("DeleteBeer", func(t *testing.T) {
		mockRepo.On("DeleteBeer", mock.Anything, testBeer.ID).Return(nil)
		err := service.DeleteBeer(context.Background(), testBeer.ID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("FilterAndPaginateBeers", func(t *testing.T) {
		mockRepo.On("FilterAndPaginateBeers", mock.Anything, testBeer.Name, int64(1), int64(10)).Return([]*model.Beer{&testBeer}, int64(1), nil)
		result, total, err := service.FilterAndPaginateBeers(context.Background(), testBeer.Name, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, []*model.Beer{&testBeer}, result)
		mockRepo.AssertExpectations(t)
	})
}
