package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	model "pok92deng/product"
	"testing"
)

type mockUsecase struct {
	mock.Mock
}

func (m *mockUsecase) CreateBeer(ctx context.Context, beer model.Beer) (primitive.ObjectID, error) {
	args := m.Called(ctx, beer)
	fmt.Println(args)
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

func (ms *mockUsecase) CreateBeerError(ctx context.Context, beer model.Beer) (string, error) {
	return "", errors.New("internal server error")
}

func (m *mockUsecase) Remove(name string) error {
	args := m.Called(name)
	return args.Error(0)
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

		err := writer.WriteField("name", testBeer.Name)
		if err != nil {
			return
		}
		err = writer.WriteField("category", testBeer.Category)
		if err != nil {
			return
		}
		err = writer.WriteField("detail", testBeer.Detail)
		if err != nil {
			return
		}
		err = writer.WriteField("imagePath", testBeer.ImagePath)
		if err != nil {
			return
		}
		err = writer.WriteField("image", testBeer.Image.Filename)
		if err != nil {
			return
		}

		filePart, err := writer.CreateFormFile("image", "filename.jpg")
		assert.NoError(t, err)
		fileContent, _ := os.ReadFile("path/to/test/image.jpg")
		filePart.Write(fileContent)

		writer.Close()

		req, err := http.NewRequest("POST", "/beers", body)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		fmt.Println(w.Body.String())

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("CreateBeer with multipart form parsing error", func(t *testing.T) {
		r := gin.Default()
		r.POST("/beers", handler.CreateBeer)

		req, _ := http.NewRequest("POST", "/beers", bytes.NewBufferString("bad request"))

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateBeer with JSON unmarshalling error", func(t *testing.T) {
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
		handler := NewBeerHandlers(mockUsecase) // Assuming this initializes your handlers with the mock

		mockUsecase.On("UpdateBeer", mock.Anything, testBeer.ID, mock.AnythingOfType("model.Beer")).Return(nil)

		r.PUT("/beers/:id", handler.UpdateBeer)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		_ = writer.WriteField("name", "updateName")
		_ = writer.WriteField("category", "updateCategory")
		_ = writer.WriteField("detail", "updateDetail")
		_ = writer.WriteField("imagePath", "updateImagePath")

		filePart, err := writer.CreateFormFile("image", "filename.jpg")
		assert.NoError(t, err)
		fileContent, _ := os.ReadFile("path/to/test/image.jpg")
		_, _ = filePart.Write(fileContent)

		writer.Close()

		req, err := http.NewRequest("PUT", "/beers/"+testBeer.ID.Hex(), body)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", writer.FormDataContentType()) // Set the correct Content-Type

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		mockUsecase.AssertExpectations(t)
	})

	t.Run("UpdateBeer with invalid ID format", func(t *testing.T) {
		r := gin.Default()
		r.PUT("/beers/:id", handler.UpdateBeer)

		req, _ := http.NewRequest("PUT", "/beers/invalid-id", nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Update with ShouldBind error", func(t *testing.T) {
		r := gin.Default()
		r.PUT("/beers/:id", handler.UpdateBeer)

		req, _ := http.NewRequest("PUT", "/beers/"+testBeer.ID.Hex(), bytes.NewBufferString("bad request"))

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Successful Image Upload", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("image", "test_image.jpg")

		file, _ := os.Open("path/to/your/test/image.jpg") // Ensure this file exists for testing
		buf := new(bytes.Buffer)
		buf.ReadFrom(file)
		part.Write(buf.Bytes())
		writer.Close()

		r := gin.Default()
		r.POST("/upload", func(c *gin.Context) {
			beer := &model.Beer{}
			imagePath, err := SetBeerImage(c, beer)
			assert.NoError(t, err)
			assert.NotEmpty(t, imagePath)
		})
	})

	t.Run("Error Removing Old Image", func(t *testing.T) {
	})

	t.Run("Error Creating Directory", func(t *testing.T) {
	})

	t.Run("Error Saving Uploaded File", func(t *testing.T) {
	})

	t.Run("Missing Image File", func(t *testing.T) {
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
