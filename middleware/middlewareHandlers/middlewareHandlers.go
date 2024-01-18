package middlewaresHandler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pok92deng/config"
	middlewares "pok92deng/middleware"
	middlewaresUsecases "pok92deng/middleware/middlewareUsecases"
	auth "pok92deng/pkg"
	"pok92deng/utils"
	"strings"
)

type IMiddlewaresHandler interface {
	JwtAuth(config middlewares.JwtAuthConfig) gin.HandlerFunc
	ParamCheck() gin.HandlerFunc
	Authorize(expectRoleId ...int) gin.HandlerFunc
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

func (h *middlewaresHandler) ParamCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, exists := c.Get("userId")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "User ID not found in context",
			})
			return
		}

		userIdStr, ok := userId.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "User ID is not of type string",
			})
			return
		}

		if c.Param("user_id") != userIdStr {
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

		if len(roles) == 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Role list is empty",
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

func (h *middlewaresHandler) JwtAuth(config middlewares.JwtAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

		var userID string
		var userRoleId int
		var isCustomer bool

		if config.AllowCustomer {
			parsedToken, err := auth.ParseCustomerToken(h.cfg.Jwt(), tokenString)
			if err == nil {
				claims := parsedToken.Claims
				userID = claims.Id.Hex() // Assuming claims.Id is a string representation of ObjectID
				userRoleId = claims.RoleId
				isCustomer = true
				c.Set("userId", userID)
				c.Set("userRoleId", userRoleId)

				if !h.middlewaresUsecase.FindAccessToken(userID, tokenString) {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"error": "no permission to access",
					})
					return
				}
			}
		}

		if !isCustomer && config.AllowAdmin {
			adminClaims, err := auth.ParseAdminToken(h.cfg.Jwt(), tokenString)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "admin token verification failed",
				})
				return
			}
			userID = adminClaims.Claims.Id.Hex() // Assuming claims.Id is a string representation of ObjectID
			userRoleId = adminClaims.Claims.RoleId
			c.Set("userId", userID)
			c.Set("userRoleId", userRoleId)
		} else if !isCustomer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "token verification failed",
			})
			return
		}

		c.Next()
	}
}
