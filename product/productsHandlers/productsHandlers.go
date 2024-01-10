package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	model "pok92deng/product"
	usecases "pok92deng/product/productsUsecases"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BeerHandler interface {
	CreateBeer(c *gin.Context)
	GetBeer(c *gin.Context)
	UpdateBeer(c *gin.Context)
	DeleteBeer(c *gin.Context)
	ListBeers(c *gin.Context)
	FilterBeersByName(c *gin.Context)
	Pagination(c *gin.Context)
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
	var beer model.Beer
	if err := c.ShouldBind(&beer); err != nil { // Changed to ShouldBind for form data
		c.JSON(http.StatusBadRequest, gin.H{"error": "Form binding error", "details": err.Error()})
		return
	}

	imagePath, err := setBeerImage(c, &beer)
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

	imagePath, err := setBeerImage(c, &beer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	beer.ImagePath = imagePath

	err = h.service.UpdateBeer(c.Request.Context(), id, beer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "beer updated"})
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

func (h *BeerHandlers) ListBeers(c *gin.Context) {
	beers, err := h.service.ListBeers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, beers)
}

func setBeerImage(c *gin.Context, beer *model.Beer) (string, error) {
	file, err := c.FormFile("image")
	if err != nil {
		return "", fmt.Errorf("failed to get form file: %w", err)
	}

	if beer.ImagePath != "" {
		oldImagePath := filepath.Join("uploads/beers", filepath.Base(beer.ImagePath))
		if err := os.Remove(oldImagePath); err != nil {
			return "", fmt.Errorf("failed to remove old image: %w", err)
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

func (h *BeerHandlers) FilterBeersByName(c *gin.Context) {
	name := c.Query("name")
	beers, err := h.service.FilterBeersByName(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, beers)
}

func (h *BeerHandlers) Pagination(c *gin.Context) {
	page, limit, err := getPaginationParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	beersData, total, err := h.service.Pagination(int64(page), int64(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve beers"})
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
		Beer:      beersData,
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
