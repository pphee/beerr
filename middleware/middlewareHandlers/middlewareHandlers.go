package middlewaresHandler

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"pok92deng/config"
	middlewaresUsecases "pok92deng/middleware/middlewareUsecases"
	auth "pok92deng/pkg"
	"pok92deng/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

type IMiddlewaresHandler interface {
	JwtAuth() gin.HandlerFunc
	ParamCheck() gin.HandlerFunc
	Authorize(expectRoleId ...int) gin.HandlerFunc
	JwtAuthAdmin() gin.HandlerFunc
}

type middlewaresHandler struct {
	cfg                config.IConfig
	middlewaresUsecase middlewaresUsecases.IMiddlewaresUsecase
}

func MiddlewaresRepository(cfg config.IConfig, middlewaresUsecase middlewaresUsecases.IMiddlewaresUsecase) IMiddlewaresHandler {
	return &middlewaresHandler{
		cfg:                cfg,
		middlewaresUsecase: middlewaresUsecase,
	}
}

func (h *middlewaresHandler) JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		result, err := auth.ParseToken(h.cfg.Jwt(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}
		claims := result.Claims

		if !h.middlewaresUsecase.FindAccessToken(claims.Id.Hex(), token) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "no permission to access",
			})
			return
		}

		c.Set("userId", claims.Id)
		c.Set("userRoleId", claims.RoleId)
		c.Next()
	}
}

func (h *middlewaresHandler) JwtAuthAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		parsedToken, err := auth.ParseAdminToken(h.cfg.Jwt(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Set("parsedToken", parsedToken)

		c.Next()
	}
}

func (h *middlewaresHandler) ParamCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, exists := c.Get("userId")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "User ID not found in context",
			})
			return
		}

		userIdStr, ok := userId.(primitive.ObjectID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "User ID is not of type ObjectID",
			})
			return
		}

		if c.Param("user_id") != userIdStr.Hex() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Parameter check failed",
			})
			return
		}
		c.Next()
	}
}

func (h *middlewaresHandler) Authorize(expectRoleId ...int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoleId, exists := c.Get("userRoleId")

		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "userRoleId not found",
			})
			return
		}

		roles, err := h.middlewaresUsecase.FindRole()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		expectValueBinary := utils.BinaryConverter(sumRoles(expectRoleId...), len(roles))
		userValueBinary := utils.BinaryConverter(userRoleId.(int), len(roles))

		for i := range userValueBinary {
			if userValueBinary[i]&expectValueBinary[i] == 1 {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "No permission to access",
		})
	}
}

// Helper function to sum roles
func sumRoles(roles ...int) int {
	sum := 0
	for _, v := range roles {
		sum += v
	}
	return sum
}
