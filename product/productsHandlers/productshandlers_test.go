package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	model "pok92deng/product"
	"testing"
)

type mockUsecase struct {
	mock.Mock
}

func (m *mockUsecase) CreateBeer(ctx context.Context, beer model.Beer) (primitive.ObjectID, error) {
	args := m.Called(ctx, beer)
	return args.Get(0).(primitive.ObjectID), args.Error(1)
}

func (m *mockUsecase) GetBeer(ctx context.Context, id primitive.ObjectID) (model.Beer, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Beer), args.Error(1)
}

func (m *mockUsecase) UpdateBeer(ctx context.Context, id primitive.ObjectID, beer model.Beer) error {
	args := m.Called(ctx, id, beer)
	return args.Error(0)
}

func (m *mockUsecase) DeleteBeer(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUsecase) FilterAndPaginateBeers(ctx context.Context, name string, page int64, limit int64) ([]*model.Beer, int64, error) {
	args := m.Called(ctx, name, page, limit)
	return args.Get(0).([]*model.Beer), args.Get(1).(int64), args.Error(2)
}

func TestBeerHandlers(t *testing.T) {
	mockUsecase := new(mockUsecase)
	handler := NewBeerHandlers(mockUsecase)

	fileHeader := &multipart.FileHeader{
		Filename: "test_beer_image.jpg",
		Header:   make(map[string][]string),
	}

	testBeer := model.Beer{
		ID:        primitive.NewObjectID(),
		Name:      "Test Beer",
		Category:  "Ale",
		Detail:    "Test Detail",
		Image:     fileHeader,
		ImagePath: "test.jpg",
	}

	t.Run("CreateBeer", func(t *testing.T) {
		r := gin.Default()
		mockUsecase.On("CreateBeer", mock.Anything, testBeer).Return(testBeer.ID, nil)

		r.POST("/beers", handler.CreateBeer)

		// Create a buffer to write our multipart form
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Add the JSON fields
		jsonPart, _ := writer.CreateFormField("json")
		jsonBytes, _ := json.Marshal(testBeer)
		jsonPart.Write(jsonBytes)

		// Add the file (if necessary)
		filePart, _ := writer.CreateFormFile("file_field_name", "filename.jpg")
		filePart.Write([]byte("file content")) // Replace with actual file content

		writer.Close()

		req, _ := http.NewRequest("POST", "/beers", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	//t.Run("GetBeer", func(t *testing.T) {
	//	r := gin.Default()
	//	mockUsecase.On("GetBeer", mock.Anything, testBeer.ID).Return(testBeer, nil)
	//
	//	r.GET("/beers/:id", handler.GetBeer)
	//
	//	req, _ := http.NewRequest("GET", "/beers/"+testBeer.ID.Hex(), nil)
	//
	//	w := httptest.NewRecorder()
	//	r.ServeHTTP(w, req)
	//	assert.Equal(t, http.StatusOK, w.Code)
	//})
	//
	//t.Run("UpdateBeer", func(t *testing.T) {
	//	r := gin.Default()
	//	mockUsecase.On("UpdateBeer", mock.Anything, testBeer.ID, testBeer).Return(nil)
	//
	//	r.PUT("/beers/:id", handler.UpdateBeer)
	//
	//	body, _ := json.Marshal(testBeer)
	//	req, _ := http.NewRequest("PUT", "/beers/"+testBeer.ID.Hex(), bytes.NewBuffer(body))
	//	req.Header.Set("Content-Type", "application/json")
	//
	//	w := httptest.NewRecorder()
	//	r.ServeHTTP(w, req)
	//	assert.Equal(t, http.StatusOK, w.Code)
	//})
	//
	//t.Run("DeleteBeer", func(t *testing.T) {
	//	r := gin.Default()
	//	mockUsecase.On("DeleteBeer", mock.Anything, testBeer.ID).Return(nil)
	//
	//	r.DELETE("/beers/:id", handler.DeleteBeer)
	//
	//	req, _ := http.NewRequest("DELETE", "/beers/"+testBeer.ID.Hex(), nil)
	//
	//	w := httptest.NewRecorder()
	//	r.ServeHTTP(w, req)
	//	assert.Equal(t, http.StatusOK, w.Code)
	//
	//})
	//
	//t.Run("FilterAndPaginateBeers", func(t *testing.T) {
	//	r := gin.Default()
	//	mockUsecase.On("FilterAndPaginateBeers", mock.Anything, testBeer.Name, int64(1), int64(10)).Return([]*model.Beer{&testBeer}, int64(1), nil)
	//
	//	r.GET("/beers", handler.FilterAndPaginateBeers)
	//
	//	req, _ := http.NewRequest("GET", "/beers", nil)
	//
	//	w := httptest.NewRecorder()
	//	r.ServeHTTP(w, req)
	//	assert.Equal(t, http.StatusOK, w.Code)
	//})
}
