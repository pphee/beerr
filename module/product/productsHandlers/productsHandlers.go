package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"pok92deng/module/product"
	"pok92deng/module/product/productsUsecases"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BeerHandler interface {
	CreateBeer(c *gin.Context)
	GetBeer(c *gin.Context)
	UpdateBeer(c *gin.Context)
	DeleteBeer(c *gin.Context)
	FilterAndPaginateBeers(c *gin.Context)
}

type BeerHandlers struct {
	service usecases.BeerService
}

func NewBeerHandlers(service usecases.BeerService) BeerHandler {
	return &BeerHandlers{
		service: service,
	}
}

func (h *BeerHandlers) CreateBeer(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form", "details": err.Error()})
		return
	}

	var beer model.Beer
	if err := c.ShouldBind(&beer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imagePath, err := SetBeerImage(c, &beer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	beer.ImagePath = imagePath

	id, err := h.service.CreateBeer(c.Request.Context(), beer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *BeerHandlers) UpdateBeer(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
		return
	}

	var beer model.Beer
	if err := c.ShouldBind(&beer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imagePath, err := SetBeerImage(c, &beer)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			imagePath = ""
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	beer.ImagePath = imagePath

	err = h.service.UpdateBeer(c.Request.Context(), id, beer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "beer updated"})
}

func (h *BeerHandlers) GetBeer(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
		return
	}

	beer, err := h.service.GetBeer(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, beer)
}

func (h *BeerHandlers) DeleteBeer(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
		return
	}

	err = h.service.DeleteBeer(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "beer deleted"})
}

type FileOps interface {
	Remove(name string) error
}

type RealFileOps struct{}

func (RealFileOps) Remove(name string) error {
	return os.Remove(name)
}

var fileOps FileOps = RealFileOps{}

func SetBeerImage(c *gin.Context, uploadImage *model.Beer) (string, error) {
	file, err := c.FormFile("image")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return uploadImage.ImagePath, nil
		}
		return "", fmt.Errorf("failed to get form file: %w", err)
	}

	if uploadImage.ImagePath != "" {
		oldImagePath := filepath.Join("uploads/beers", filepath.Base(uploadImage.ImagePath))
		if _, err := os.Stat(oldImagePath); err == nil {
			if err := fileOps.Remove(oldImagePath); err != nil {
				return "", fmt.Errorf("failed to remove old image: %w", err)
			}
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("error checking for old image: %w", err)
		}
	}

	dirPath := filepath.Join("uploads/beers")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	filePath := filepath.Join(dirPath, file.Filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		return "", fmt.Errorf("failed to save uploaded file: %w", err)
	}

	return "https://hzbxs242-3000.asse.devtunnels.ms" + "/" + filePath, nil
}

func (h *BeerHandlers) FilterAndPaginateBeers(c *gin.Context) {
	name := c.Query("name")
	page, limit, err := getPaginationParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	beers, total, err := h.service.FilterAndPaginateBeers(c.Request.Context(), name, int64(page), int64(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pagination, err := getPagination(c, int(total))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := &model.BeerPagingResult{
		Page:      pagination.Page,
		Limit:     pagination.Limit,
		PrevPage:  pagination.PrevPage,
		NextPage:  pagination.NextPage,
		Count:     pagination.Count,
		TotalPage: pagination.TotalPage,
		Data:      beers,
	}

	c.JSON(http.StatusOK, response)
}

func getPaginationParams(c *gin.Context) (page, limit int, err error) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page, err = strconv.Atoi(pageStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid page number")
	}
	limit, err = strconv.Atoi(limitStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid limit number")
	}

	return page, limit, nil
}

func getPagination(c *gin.Context, total int) (*model.BeerPagingResult, error) {
	page, limit, err := getPaginationParams(c)
	if err != nil {
		return nil, err
	}
	totalPages := (total + limit - 1) / limit

	return &model.BeerPagingResult{
		Page:      page,
		Limit:     limit,
		PrevPage:  maxInt(1, page-1),
		NextPage:  minInt(totalPages, page+1),
		Count:     total,
		TotalPage: totalPages,
	}, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

//func parseMultipartForm(c *gin.Context) func() mo.Either[error, *gin.Context] {
//	return func() mo.Either[error, *gin.Context] {
//		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
//			return mo.NewLeft(fmt.Errorf("failed to parse multipart form: %w", err))
//		}
//		return mo.NewRight(c)
//	}
//}
//
//func bindBeer(c *gin.Context) func() mo.Either[error, model.Beer] {
//	return func() mo.Either[error, model.Beer] {
//		var beer model.Beer
//		if err := c.ShouldBind(&beer); err != nil {
//			return mo.NewLeft(fmt.Errorf("failed to bind request: %w", err))
//		}
//		return mo.NewRight(beer)
//	}
//}
//
//func setBeerImageFunc(beer model.Beer) func(c *gin.Context) mo.Either[error, model.Beer] {
//	return func(c *gin.Context) mo.Either[error, model.Beer] {
//		imagePath, err := handleSetBeerImage(c, &beer)
//		if err != nil {
//			return mo.NewLeft(fmt.Errorf("failed to set beer image: %w", err))
//		}
//		beer.ImagePath = imagePath
//		return mo.NewRight(beer)
//	}
//}
//
//func createBeerFunc(beer model.Beer) func(c *gin.Context, service usecases.BeerService) mo.Either[error, string] {
//	return func(c *gin.Context, service usecases.BeerService) mo.Either[error, string] {
//		id, err := service.CreateBeer(c.Request.Context(), beer)
//		if err != nil {
//			return mo.NewLeft(fmt.Errorf("failed to create beer: %w", err))
//		}
//		return mo.NewRight(id)
//	}
//}
//
//func createBeerMonad(c *gin.Context, service usecases.BeerService) mo.Either[error, string] {
//	return fp.Pipe2(
//		parseMultipartForm(c),
//		bindBeer(c),
//		func(beer model.Beer) mo.Either[error, model.Beer] {
//			return setBeerImageFunc(beer)(c)
//		},
//		func(beer model.Beer) mo.Either[error, string] {
//			return createBeerFunc(beer)(c, service)
//		},
//	)
//}
//
//func (h *BeerHandlers) CreateBeer(c *gin.Context) {
//	result := createBeerMonad(c, h.service)
//	result.Match(
//		func(id string) {
//			c.JSON(http.StatusCreated, gin.H{"id": id})
//		},
//		func(err error) {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		},
//	)
//}

//func parseMultipartForm(c *gin.Context) func() mo.Either[error, *gin.Context] {
//	return func() mo.Either[error, *gin.Context] {
//		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
//			return mo.NewLeft(fmt.Errorf("failed to parse multipart form: %w", err))
//		}
//		return mo.NewRight(c)
//	}
//}
//
//func bindBeer(c *gin.Context) func() mo.Either[error, model.Beer] {
//	return func() mo.Either[error, model.Beer] {
//		var beer model.Beer
//		if err := c.ShouldBind(&beer); err != nil {
//			return mo.NewLeft(fmt.Errorf("failed to bind request: %w", err))
//		}
//		return mo.NewRight(beer)
//	}
//}
//
//func setBeerImageFunc(beer model.Beer) func(c *gin.Context) mo.Either[error, model.Beer] {
//	return func(c *gin.Context) mo.Either[error, model.Beer] {
//		imagePath, err := handleSetBeerImage(c, &beer)
//		if err != nil {
//			return mo.NewLeft(fmt.Errorf("failed to set beer image: %w", err))
//		}
//		beer.ImagePath = imagePath
//		return mo.NewRight(beer)
//	}
//}
//
//func createBeerFunc(beer model.Beer) func(c *gin.Context, service usecases.BeerService) mo.Either[error, string] {
//	return func(c *gin.Context, service usecases.BeerService) mo.Either[error, string] {
//		id, err := service.CreateBeer(c.Request.Context(), beer)
//		if err != nil {
//			return mo.NewLeft(fmt.Errorf("failed to create beer: %w", err))
//		}
//		return mo.NewRight(id)
//	}
//}
//
//func createBeerMonad(c *gin.Context, service usecases.BeerService) mo.Either[error, string] {
//	pipeline := fp.Pipe2(
//		parseMultipartForm(c),
//		bindBeer(c),
//		setBeerImageFunc,
//		createBeerFunc,
//	)
//
//	return pipeline(c, service)
//}
//
//func (h *BeerHandlers) CreateBeer(c *gin.Context) {
//	result := createBeerMonad(c, h.service)
//	result.Match(
//		func(id string) {
//			c.JSON(http.StatusCreated, gin.H{"id": id})
//		},
//		func(err error) {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		},
//	)
//}

//func checkImagePathEmpty(uploadImage *model.Beer) mo.Either[error, *model.Beer] {
//	if uploadImage.ImagePath == "" {
//		return mo.NewLeft(fmt.Errorf("image path is empty")) // Error: image path is empty
//	}
//	return mo.NewRight(uploadImage) // Success: image path is not empty
//}
//
//func checkForOldImage(uploadImage *model.Beer) mo.Either[error, *model.Beer] {
//	oldImagePath := filepath.Join("uploads/beers", filepath.Base(uploadImage.ImagePath))
//
//	if _, err := os.Stat(oldImagePath); os.IsNotExist(err) {
//		return mo.NewRight(uploadImage) // Success: old image doesn't exist
//	}
//	return mo.NewLeft(fmt.Errorf("old image exists at path: %s", oldImagePath)) // Error: old image exists
//}
//
//func removeOldImage(uploadImage *model.Beer) mo.Either[error, *model.Beer] {
//	oldImagePath := filepath.Join("uploads/beers", filepath.Base(uploadImage.ImagePath))
//
//	if err := fileOps.Remove(oldImagePath); err != nil {
//		return mo.NewLeft(fmt.Errorf("failed to remove old image: %w", err)) // Error: failed to remove
//	}
//	return mo.NewRight(uploadImage) // Success: old image removed
//}
