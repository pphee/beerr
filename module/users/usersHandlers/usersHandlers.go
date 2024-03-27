package usersHandlers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/http/middleware"
	"golang.org/x/exp/slog"
	"net/http"
	"pok92deng/config"
	"pok92deng/module/users"
	"pok92deng/module/users/usersUsecases"
	"strconv"
	"strings"
)

type IUserHandler interface {
	SignUpCustomer(c *gin.Context)
	SignIn(c *gin.Context)
	RefreshPassport(c *gin.Context)
	GerUserProfile(c *gin.Context)
	SignUpAdmin(c *gin.Context)
	RefreshPassportAdmin(c *gin.Context)
	GetAllUserProfile(c *gin.Context)
	UpdateRole(c *gin.Context)
	CreateRole(c *gin.Context)
	AuthCtx(c *gin.Context)
	CreateUserZitadel(c *gin.Context)
}

type usersHandler struct {
	usersUsecase usersUsecases.IUsersUsecase
	cfg          config.IConfig
	mw           *middleware.Interceptor[*oauth.IntrospectionContext]
}

func UsersHandler(cfg config.IConfig, usersUsecase usersUsecases.IUsersUsecase, mw *middleware.Interceptor[*oauth.IntrospectionContext]) IUserHandler {
	return &usersHandler{
		cfg:          cfg,
		usersUsecase: usersUsecase,
		mw:           mw,
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

	//Email validation
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
		fmt.Println(err)
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

//func (h *usersHandler) GenerateAdminToken(c *gin.Context) {
//	adminToken, err := auth.NewAuth(auth.Admin, h.cfg.Jwt(), nil)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{
//			"error":   "Internal Server Error",
//			"message": err.Error(),
//		})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"token": adminToken.SignToken(),
//	})
//}

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

func (h *usersHandler) GetAllUserProfile(c *gin.Context) {
	result, err := h.usersUsecase.GetAllUserProfile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *usersHandler) UpdateRole(c *gin.Context) {
	userId := strings.TrimSpace(c.Param("user_id"))

	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var req users.RoleUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.RoleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role ID is required"})
		return
	}

	roleId, err := strconv.Atoi(req.RoleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	if req.Role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role is required"})
		return
	}

	if err := h.usersUsecase.UpdateRole(userId, roleId, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Update role successfully"})
}

func (h *usersHandler) CreateRole(c *gin.Context) {
	var req users.RoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	if err := h.usersUsecase.CreateRole(req.RoleID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Create role successfully",
	})
}

func (h *usersHandler) ginMw(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.mw.RequireAuthorization()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
			handler(c)
		})).ServeHTTP(c.Writer, c.Request)
	}
}

func (h *usersHandler) AuthCtx(c *gin.Context) {
	h.ginMw(func(c *gin.Context) {
		authCtx := h.mw.Context(c.Request.Context())
		if authCtx == nil {
			slog.Error("failed to get authorization context", "error", "authCtx is nil")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get authorization context"})
			return
		}

		userID := authCtx.UserID()
		userEmail := authCtx.Email

		slog.Info("user accessed task list", "id", userID, "username", authCtx.Username)
		slog.Info("user accessed task list", "email", userEmail)

		userCredential := &users.UserCredential{
			Email: userEmail,
		}

		fmt.Println("userCredential", userCredential)

		passport, err := h.usersUsecase.GetPassport(userCredential)
		if err != nil {
			slog.Error("failed to get user passport", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user passport"})
			return
		}

		zitadel := &users.UserPassport{
			User: &users.User{
				Id:       passport.User.Id,
				Email:    passport.User.Email,
				Username: passport.User.Username,
				Role:     passport.User.Role,
				RoleId:   passport.User.RoleId,
			},
		}

		c.JSON(http.StatusOK, zitadel)

	})(c)
}

func (h *usersHandler) CreateUserZitadel(c *gin.Context) {
	req := new(users.UserRegisterReq)
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": err.Error(),
		})
		return
	}
	ctx := c.Request.Context()

	result, err := h.usersUsecase.CreateUserZitadel(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"data":    result,
	})
}
