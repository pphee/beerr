package middlewaresHandler

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/http/middleware"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
	"golang.org/x/exp/slog"
	"net/http"
	"pok92deng/config"
	"pok92deng/module/middleware"
	"pok92deng/module/middleware/middlewareUsecases"
	"pok92deng/module/users"
	"pok92deng/module/users/usersUsecases"
	auth "pok92deng/pkg"
	"pok92deng/pkg/utils"
	"strconv"
	"strings"
)

type IMiddlewaresHandler interface {
	JwtAuth(config middlewares.JwtAuthConfig) gin.HandlerFunc
	ParamCheck() gin.HandlerFunc
	Authorize(expectRoleId ...int) gin.HandlerFunc
	AuthorizeString(expectRoleNames ...middlewares.UsersRole) gin.HandlerFunc
}

type middlewaresHandler struct {
	cfg                config.IConfig
	middlewaresUsecase middlewaresUsecases.IMiddlewaresUsecase
	usersUsecase       usersUsecases.IUsersUsecase
}

func MiddlewaresHandler(cfg config.IConfig, middlewaresUsecase middlewaresUsecases.IMiddlewaresUsecase, usersUsecase usersUsecases.IUsersUsecase) IMiddlewaresHandler {
	return &middlewaresHandler{
		cfg:                cfg,
		middlewaresUsecase: middlewaresUsecase,
		usersUsecase:       usersUsecase,
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
		fmt.Println("userRoleId", userRoleId, "exists", exists)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "userRoleId not found"})
			return
		}

		userRoleIdInt, ok := userRoleId.(int)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "userRoleId is of an invalid type"})
			return
		}

		userRoleIdString := strconv.Itoa(userRoleIdInt)
		roles, err := h.middlewaresUsecase.FindRole(c.Request.Context(), userRoleIdString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		isOk := false
		for _, roleId := range expectRoleId {
			if roleExists(roleId, roles) {
				isOk = true
				break
			}
		}

		if !isOk {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No required role found"})
			return
		}
		expectValueBinary := utils.BinaryConverter(sumRoles(expectRoleId...), 10)
		userValueBinary := utils.BinaryConverter(userRoleIdInt, 10)

		for i := range userValueBinary {
			if userValueBinary[i]&expectValueBinary[i] == 1 {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No permission to access"})
	}
}

func roleExists(roleId int, roles []*middlewares.Roles) bool {
	for _, role := range roles {
		if role == nil {
			continue
		}
		parsedRoleId, err := strconv.Atoi(role.RoleID)
		if err != nil {
			fmt.Println("Error parsing role ID:", role.RoleID, "Error:", err)
			continue
		}
		if parsedRoleId == roleId {
			return true
		}

	}
	return false

}

func sumRoles(roles ...int) int {
	sum := 0
	for _, v := range roles {
		sum += v
	}
	return sum
}

func (h *middlewaresHandler) JwtAuth(config middlewares.JwtAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		loginType := c.Request.Header.Get("loginType")
		fmt.Println("loginType", loginType)
		switch loginType {
		case "normal":
			h.processToken(c, config)
		case "zitadel":
			h.processZitadelToken(c)
		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "loginType not found"})
		}
	}
}

func (h *middlewaresHandler) processZitadelToken(c *gin.Context) {
	ctx := context.Background()
	authZ, err := authorization.New(ctx, zitadel.New(h.cfg.Zitadel().Domain()), oauth.DefaultAuthorization(h.cfg.Zitadel().Key()))
	if err != nil {
		slog.Error("zitadel sdk could not initialize", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Zitadel SDK initialization failed"})
		return
	}
	mw := middleware.New(authZ)

	var userEmail string
	var userID string

	adaptedMiddleware := func(c *gin.Context) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(r.Context())
			authCtx := mw.Context(r.Context())
			if authCtx == nil {
				slog.Error("failed to get authorization context", "error", "authCtx is nil")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get authorization context"})
				return
			}

			userID = authCtx.UserID()
			userEmail = authCtx.Email

			slog.Info("user accessed task list", "id", userID, "username", authCtx.Username)
			slog.Info("user accessed task list", "email", userEmail)
		})

		mw.RequireAuthorization()(handler).ServeHTTP(c.Writer, c.Request)
	}

	adaptedMiddleware(c)

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

	roleId := passport.User.RoleId
	c.Set("userId", passport.User.Id)
	c.Set("userRoleId", roleId)

	c.Next()
}

func (h *middlewaresHandler) processToken(c *gin.Context, config middlewares.JwtAuthConfig) {
	tokenString := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

	var userID string
	var userRoleId int
	var isCustomer bool

	if config.AllowCustomer {
		parsedToken, err := auth.ParseCustomerToken(h.cfg.Jwt(), tokenString)
		if err == nil {
			claims := parsedToken.Claims
			fmt.Println("claims", claims)
			userID = claims.Id.Hex()
			userRoleId = claims.RoleId
			isCustomer = true
			c.Set("userId", userID)
			c.Set("userRoleId", userRoleId)

			if !h.middlewaresUsecase.FindAccessToken(userID, tokenString) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no permission to access"})
				return
			}
		}
	}

	if !isCustomer && config.AllowAdmin {
		adminClaims, err := auth.ParseAdminToken(h.cfg.Jwt(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "admin token verification failed"})
			return
		}
		fmt.Println("adminClaims", adminClaims.Claims)
		userID = adminClaims.Claims.Id.Hex()
		userRoleId = adminClaims.Claims.RoleId
		c.Set("userId", userID)
		c.Set("userRoleId", userRoleId)
	} else if !isCustomer {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token verification failed"})
		return
	}

	c.Next()
}

func (h *middlewaresHandler) AuthorizeString(expectRoleNames ...middlewares.UsersRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoleIdInterface, exists := c.Get("userRoleId")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "userRoleId not found"})
			return
		}

		userRoleId, ok := userRoleIdInterface.(int)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid userRoleId format"})
			return
		}

		rolesMap := map[middlewares.UsersRole]int{
			"user":    1,
			"admin":   2,
			"manager": 4,
		}

		expectedRoleBinaries := getRoleBinaries(rolesMap, expectRoleNames...)
		userValueBinary := utils.BinaryConverter(userRoleId, len(rolesMap))

		for _, roleBinary := range expectedRoleBinaries {
			if isEqual(userValueBinary, roleBinary) {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No permission to access"})
	}
}

func getRoleBinaries(rolesMap map[middlewares.UsersRole]int, roles ...middlewares.UsersRole) [][]int {
	var binaries [][]int
	for _, roleName := range roles {
		if roleValue, exists := rolesMap[roleName]; exists {
			binaries = append(binaries, utils.BinaryConverter(roleValue, len(rolesMap)))
		}
	}
	return binaries
}

func isEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
