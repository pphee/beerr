package middlewaresHandler

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"net/http/httptest"
	"pok92deng/config"
	"pok92deng/module/middleware"
	"pok92deng/module/users"
	auth "pok92deng/pkg"
	"testing"
)

type mockUsecase struct {
	mock.Mock
}

func (m *mockUsecase) FindRole(ctx context.Context, userRoleId string) ([]*middlewares.Roles, error) {
	args := m.Called(ctx, userRoleId)
	return args.Get(0).([]*middlewares.Roles), args.Error(1)
}

func (m *mockUsecase) FindAccessToken(userId, accessToken string) bool {
	args := m.Called(userId, accessToken)
	return args.Bool(0)
}

type MockAuth struct {
	mock.Mock
}

func (m *MockAuth) ParseCustomerToken(tokenString string) (*auth.MapClaims, error) {
	args := m.Called(tokenString)
	return args.Get(0).(*auth.MapClaims), args.Error(1)
}

func (m *MockAuth) ParseAdminToken(tokenString string) (*auth.MapClaims, error) {
	args := m.Called(tokenString)
	return args.Get(0).(*auth.MapClaims), args.Error(1)
}

func makeConfig() config.IConfig {
	cfg := config.LoadConfig("/Users/p/Goland/pok92deng/.env")
	return cfg
}

var someError = errors.New("invalid token error")

func TestMiddlewaresHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	cfg := makeConfig()
	mockUsecase := new(mockUsecase)
	mockAuth := new(MockAuth)
	handler := MiddlewaresRepository(cfg, mockUsecase)
	router.Use(handler.ParamCheck())
	router.Use(handler.Authorize(1))

	router.GET("/test", handler.Authorize(2), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	router.Use(handler.JwtAuth(middlewares.JwtAuthConfig{
		AllowCustomer: true,
		AllowAdmin:    true,
	}))

	t.Run("UserIdNotInContext", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/some-route", nil)
		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "User ID not found in context")
	})

	t.Run("userId Not String", func(t *testing.T) {
		router := gin.Default()
		router.Use(func(c *gin.Context) {
			c.Set("userId", 123) // Setting a non-string value
			c.Next()
		})
		router.Use((&middlewaresHandler{}).ParamCheck())
		router.GET("/test/:user_id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test/123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "User ID is not of type string")
	})

	t.Run("userId Mismatch", func(t *testing.T) {
		router := gin.Default()
		router.Use(func(c *gin.Context) {
			c.Set("userId", "456")
			c.Next()
		})
		router.Use((&middlewaresHandler{}).ParamCheck())
		router.GET("/test/:user_id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test/123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Parameter check failed")
	})

	t.Run("UserRoleIDNotFound", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/test", nil)

		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "{\"error\":\"User ID not found in context\"}")
	})

	t.Run("userId Match", func(t *testing.T) {
		router := gin.Default()
		router.Use(func(c *gin.Context) {
			c.Set("userId", "123")
			c.Next()
		})
		router.Use((&middlewaresHandler{}).ParamCheck())
		router.GET("/test/:user_id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test/123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	})

	t.Run("userId Not Present", func(t *testing.T) {
		router := gin.Default()
		router.Use((&middlewaresHandler{}).ParamCheck())
		router.GET("/test/:user_id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test/123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "User ID not found in context")
	})

	t.Run("No Permission", func(t *testing.T) {
		router := gin.Default()
		someRoles := []*middlewares.Role{
			{Id: 2, Title: "admin"},
			{Id: 1, Title: "user"},
		}
		mockUsecase.On("FindRole").Return(someRoles, nil)
		fmt.Println("someRoles", someRoles)

		router.Use(func(c *gin.Context) {
			c.Set("userRoleId", 2)
			c.Next()
		})
		router.GET("/test", handler.Authorize(1), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "No permission to access")
	})

	t.Run("Has Permission", func(t *testing.T) {
		router := gin.Default()
		someRoles := []*middlewares.Role{
			{Id: 2, Title: "admin"},
			{Id: 1, Title: "user"},
		}
		mockUsecase.On("FindRole").Return(someRoles, nil)

		router.Use(func(c *gin.Context) {
			c.Set("userRoleId", 1)
			c.Next()
		})
		router.GET("/test", handler.Authorize(1), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	})

	t.Run("userRoleId_Not_Present", func(t *testing.T) {
		router := gin.Default()

		var someRoles []*middlewares.Role
		mockUsecase.On("FindRole").Return(someRoles, nil)

		router.GET("/test", handler.Authorize(1), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "userRoleId not found")
	})

	t.Run("Valid Customer Token", func(t *testing.T) {
		router := gin.Default()
		router.Use(handler.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: true}))
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		mockToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGFpbXMiOnsiaWQiOiI2NTllNTc5M2NkZmExMzFlMDM5YTRhOTMiLCJyb2xlIjoxfSwiaXNzIjoiZ29FY29tbWVyY2UtYXBpIiwic3ViIjoiYWNjZXNzLXRva2VuIiwiYXVkIjpbInVzZXIiXSwiZXhwIjoxNzA1NjU4NDA4LCJuYmYiOjE3MDU1NzIwMDgsImlhdCI6MTcwNTU3MjAwOH0.Vjgex4mKXnMPpNl1AX4vuu5qSzJEE4qswr9zSDapn7s"

		mockAuth.On("ParseCustomerToken", mockToken).Return(&auth.MapClaims{
			Claims: &users.UserClaims{
				Id:     primitive.NewObjectID(),
				RoleId: 1,
			},
		}, nil)

		mockUsecase.On("FindAccessToken", "659e5793cdfa131e039a4a93", mockToken).Return(true)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+mockToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Valid Admin Token", func(t *testing.T) {
		router := gin.Default()
		router.Use(handler.JwtAuth(middlewares.JwtAuthConfig{AllowAdmin: true}))
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		mockToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGFpbXMiOnsiaWQiOiI2NTk3ZDNkYjhlYjI4ZjM4ZDY1Y2M5YWIiLCJyb2xlIjoyfSwiaXNzIjoiZ29FY29tbWVyY2UtYXBpIiwic3ViIjoiYWRtaW4tdG9rZW4iLCJhdWQiOlsiYWRtaW4iXSwiZXhwIjoxNzA1NTcyMjU1LCJuYmYiOjE3MDU1NzE5NTUsImlhdCI6MTcwNTU3MTk1NX0.UB4oXQPDV4mHaeP3D6vzdhVRX1qwL3XloJRqkauHHSM"

		mockAuth.On("ParseAdminToken", mockToken).Return(&auth.MapClaims{
			Claims: &users.UserClaims{
				Id:     primitive.NewObjectID(), // Use the ID that's expected in the mock token
				RoleId: 2,
			},
		}, nil)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+mockToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

	})

}
