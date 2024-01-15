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
	"strings"
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
	gin.SetMode(gin.TestMode)

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
		mockUsecase.On("CreateBeer", mock.Anything, mock.AnythingOfType("model.Beer")).Return(testBeer.ID, nil)

		r.POST("/beers", handler.CreateBeer)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		jsonPart, err := writer.CreateFormField("beer")
		assert.NoError(t, err)

		jsonBytes, err := json.Marshal(testBeer)
		assert.NoError(t, err)
		jsonPart.Write(jsonBytes)

		filePart, err := writer.CreateFormFile("image", "filename.jpg")
		assert.NoError(t, err)
		filePart.Write([]byte("file content"))

		writer.Close()

		req, err := http.NewRequest("POST", "/beers", body)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("CreateBeer with multipart form parsing error", func(t *testing.T) {
		// Setup
		r := gin.Default()
		r.POST("/beers", handler.CreateBeer)

		// Sending a bad request
		req, _ := http.NewRequest("POST", "/beers", bytes.NewBufferString("bad request"))

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateBeer with JSON unmarshalling error", func(t *testing.T) {
		// Setup
		r := gin.Default()
		r.POST("/beers", handler.CreateBeer)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		jsonPart, _ := writer.CreateFormField("beer")
		jsonPart.Write([]byte("{malformed JSON"))

		writer.Close()
		req, _ := http.NewRequest("POST", "/beers", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateBeer with internal server error", func(t *testing.T) {
		r := gin.Default()
		r.POST("/beers", handler.CreateBeer)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		jsonPart, _ := writer.CreateFormField("beer")
		jsonBytes, _ := json.Marshal(testBeer)
		jsonPart.Write(jsonBytes)

		writer.Close()
		req, _ := http.NewRequest("POST", "/beers", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

	})

	t.Run("GetBeer", func(t *testing.T) {
		r := gin.Default()
		mockUsecase.On("GetBeer", mock.Anything, testBeer.ID).Return(testBeer, nil)

		r.GET("/beers/:id", handler.GetBeer)

		req, _ := http.NewRequest("GET", "/beers/"+testBeer.ID.Hex(), nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetBeer with invalid ID format", func(t *testing.T) {
		r := gin.Default()
		r.GET("/beers/:id", handler.GetBeer)

		req, _ := http.NewRequest("GET", "/beers/invalid-id", nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UpdateBeer", func(t *testing.T) {
		r := gin.Default()
		mockUsecase.On("UpdateBeer", mock.Anything, testBeer.ID, testBeer).Return(nil)

		r.PUT("/beers/:id", handler.UpdateBeer)

		body, _ := json.Marshal(testBeer)
		req, _ := http.NewRequest("PUT", "/beers/"+testBeer.ID.Hex(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("UpdateBeer with invalid ID format", func(t *testing.T) {
		r := gin.Default()
		r.PUT("/beers/:id", handler.UpdateBeer)

		req, _ := http.NewRequest("PUT", "/beers/invalid-id", nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UpdateBeer with binding error", func(t *testing.T) {
		r := gin.Default()
		r.PUT("/beers/:id", handler.UpdateBeer)

		incompleteData := `name=Test&detail=Test Detail`
		req, _ := http.NewRequest("PUT", "/beers/valid-id", strings.NewReader(incompleteData))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("FilterAndPaginateBeers", func(t *testing.T) {
		r := gin.Default()
		mockUsecase.On("FilterAndPaginateBeers", mock.Anything, testBeer.Name, int64(1), int64(10)).Return([]*model.Beer{&testBeer}, int64(1), nil)

		r.GET("/beers", handler.FilterAndPaginateBeers)

		req, _ := http.NewRequest("GET", "/beers?name="+testBeer.Name+"&page=1&limit=10", nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("FilterAndPaginateBeers with invalid page", func(t *testing.T) {
		r := gin.Default()
		r.GET("/beers", handler.FilterAndPaginateBeers)

		req, _ := http.NewRequest("GET", "/beers?name="+testBeer.Name+"&page=invalid-page&limit=10", nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("FilterAndPaginateBeers with invalid limit", func(t *testing.T) {
		r := gin.Default()
		r.GET("/beers", handler.FilterAndPaginateBeers)

		req, _ := http.NewRequest("GET", "/beers?name="+testBeer.Name+"&page=1&limit=invalid-limit", nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("FilterAndPaginateBeers with error", func(t *testing.T) {
		r := gin.Default()
		mockUsecase.On("FilterAndPaginateBeers", mock.Anything, testBeer.Name, int64(1), int64(10)).Return([]*model.Beer{&testBeer}, int64(1), nil)

		r.GET("/beers", handler.FilterAndPaginateBeers)

		req, _ := http.NewRequest("GET", "/beers?name="+testBeer.Name+"&page=1&limit=10", nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("FilterAndPaginateBeers with invalid pagination parameters", func(t *testing.T) {
		r := gin.Default()
		r.GET("/beers", handler.FilterAndPaginateBeers)

		req, _ := http.NewRequest("GET", "/beers?page=invalid&limit=invalid", nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("DeleteBeer", func(t *testing.T) {
		r := gin.Default()
		mockUsecase.On("DeleteBeer", mock.Anything, testBeer.ID).Return(nil)

		r.DELETE("/beers/:id", handler.DeleteBeer)

		req, _ := http.NewRequest("DELETE", "/beers/"+testBeer.ID.Hex(), nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

	})

	t.Run("DeleteBeer with invalid ID format", func(t *testing.T) {
		r := gin.Default()
		r.DELETE("/beers/:id", handler.DeleteBeer)

		req, _ := http.NewRequest("DELETE", "/beers/invalid-id", nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
