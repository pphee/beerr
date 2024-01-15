package usersHandlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"pok92deng/config"
	auth "pok92deng/pkg"
	"pok92deng/users"
	"pok92deng/users/usersUsecases"
	"strings"
)

type IUserHandler interface {
	SignUpCustomer(c *gin.Context)
	SignIn(c *gin.Context)
	RefreshPassport(c *gin.Context)
	GenerateAdminToken(c *gin.Context)
	GerUserProfile(c *gin.Context)
	SignUpAdmin(c *gin.Context)
	RefreshPassportAdmin(c *gin.Context)
}

type usersHandler struct {
	usersUsecase usersUsecases.IUsersUsecase
	cfg          config.IConfig
}

func UsersHandler(cfg config.IConfig, usersUsecase usersUsecases.IUsersUsecase) IUserHandler {
	return &usersHandler{
		cfg:          cfg,
		usersUsecase: usersUsecase,
	}
}

func (h *usersHandler) SignUpCustomer(c *gin.Context) {
	// Request body parser
	req := new(users.UserRegisterReq)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "Invalid request",
		})
		return
	}

	// Email validation
	if !req.IsEmail() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid email format",
			"message": "email pattern is invalid",
		})
		return
	}

	//if !req.IsPassword() {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"error":   "Invalid password format",
	//		"message": "password pattern is invalid",
	//	})
	//	return
	//}

	// Insert
	result, err := h.usersUsecase.InsertCustomer(req)
	if err != nil {
		var httpErr *users.HTTPError
		if errors.As(err, &httpErr) {
			c.JSON(httpErr.StatusCode, gin.H{
				"error": httpErr.Message,
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Customer created successfully",
		"result":  result,
	})
}

func (h *usersHandler) SignIn(c *gin.Context) {
	var req users.UserCredential

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	passport, err := h.usersUsecase.GetPassport(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, passport)
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (h *usersHandler) RefreshPassport(c *gin.Context) {
	req := new(users.UserRefreshCredential)
	if err := c.ShouldBindJSON(req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	passport, err := h.usersUsecase.RefreshPassport(req)
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Error refreshing passport: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, passport)
}

func (h *usersHandler) RefreshPassportAdmin(c *gin.Context) {
	req := new(users.UserRefreshCredential)
	if err := c.ShouldBindJSON(req); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	passport, err := h.usersUsecase.RefreshPassportAdmin(req)
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Error refreshing passport: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, passport)
}

func (h *usersHandler) respondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, ErrorResponse{Code: code, Message: message})
}

func (h *usersHandler) SignUpAdmin(c *gin.Context) {
	req := new(users.UserRegisterReq)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": err.Error(),
		})
		return
	}

	// Email validation
	if !req.IsEmail() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "email pattern is invalid",
		})
		return
	}

	// Insert
	result, err := h.usersUsecase.InsertAdmin(req)
	if err != nil {
		var httpErr *users.HTTPError
		if errors.As(err, &httpErr) {
			c.JSON(httpErr.StatusCode, gin.H{
				"error": httpErr.Message,
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Success",
		"data":    result,
	})
}

func (h *usersHandler) GenerateAdminToken(c *gin.Context) {
	adminToken, err := auth.NewAuth(auth.Admin, h.cfg.Jwt(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": adminToken.SignToken(),
	})
}

func (h *usersHandler) GerUserProfile(c *gin.Context) {
	userId := strings.Trim(c.Param("user_id"), " ")

	// Get profile
	result, err := h.usersUsecase.GetUserProfile(userId)
	if err != nil {
		switch err.Error() {
		case "get user failed: mongodb: no rows in result set":
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK, result)
}
