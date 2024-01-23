package usersHandlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"net/http/httptest"
	"pok92deng/config"
	"pok92deng/users"
	"strings"
	"testing"
)

type mockUsecases struct {
	mock.Mock
}

func (m *mockUsecases) CreateRole(roleId, role string) error {
	args := m.Called(roleId, role)
	return args.Error(0)
}

func (m *mockUsecases) GetAllUserProfile() ([]*users.User, error) {
	args := m.Called()
	return args.Get(0).([]*users.User), args.Error(1)
}

func (m *mockUsecases) UpdateRole(userId string, roleId int, role string) error {
	args := m.Called(userId, roleId, role)
	return args.Error(0)
}

func (m *mockUsecases) RefreshPassportAdmin(req *users.UserRefreshCredential) (*users.UserPassport, error) {
	args := m.Called(req)
	return args.Get(0).(*users.UserPassport), args.Error(1)
}

func (m *mockUsecases) InsertCustomer(req *users.UserRegisterReq) (*users.UserPassport, error) {
	args := m.Called(req)
	return args.Get(0).(*users.UserPassport), args.Error(1)
}

func (m *mockUsecases) GetPassport(req *users.UserCredential) (*users.UserPassport, error) {
	args := m.Called(req)
	return args.Get(0).(*users.UserPassport), args.Error(1)
}

func (m *mockUsecases) RefreshPassport(req *users.UserRefreshCredential) (*users.UserPassport, error) {
	args := m.Called(req)
	return args.Get(0).(*users.UserPassport), args.Error(1)
}

func (m *mockUsecases) GetUserProfile(userId string) (*users.User, error) {
	args := m.Called(userId)
	return args.Get(0).(*users.User), args.Error(1)
}

func (m *mockUsecases) InsertAdmin(req *users.UserRegisterReq) (*users.UserPassport, error) {
	args := m.Called(req)
	return args.Get(0).(*users.UserPassport), args.Error(1)
}

func makeConfig() config.IConfig {
	cfg := config.LoadConfig("/Users/p/Goland/pok92deng/.env")
	return cfg
}

func TestSignUpCustomer(t *testing.T) {

	testCustomer := &users.UserRegisterReq{
		Email:    "pok92deng@example.com",
		Password: "Complex!@3",
		Username: "pok92deng",
	}

	validCredentials := &users.UserCredential{
		Email:    "pok92deng@example.com",
		Password: "Complex!@3",
	}

	validRefreshCredential := &users.UserRefreshCredential{
		RefreshToken: "validRefreshToken",
	}

	t.Run("SignUpCustomer", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)
		mockUsecases.On("InsertCustomer", testCustomer).Return(&users.UserPassport{}, nil)

		r := gin.Default()
		r.POST("/signup", handler.SignUpCustomer)

		reqBody, _ := json.Marshal(testCustomer)
		req, _ := http.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusCreated, resp.Code)

		var responseBody map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "Customer created successfully", responseBody["message"])

		mockUsecases.AssertExpectations(t)
	})

	t.Run("Successful Insert Admin", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)
		testAdmin := &users.UserRegisterReq{
			Email:    "admin@example.com",
			Password: "adminpass",
			Username: "adminuser",
		}

		mockUsecases.On("InsertAdmin", testAdmin).Return(&users.UserPassport{
			User: &users.User{
				Id:       primitive.NewObjectID(),
				Email:    testAdmin.Email,
				Username: testAdmin.Username,
				Role:     "admin",
				RoleId:   2,
			},
		}, nil)

		r := gin.Default()
		r.POST("/signup-admin", handler.SignUpAdmin)

		reqBody, _ := json.Marshal(testAdmin)
		req, _ := http.NewRequest(http.MethodPost, "/signup-admin", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusCreated, resp.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)
		r := gin.Default()
		r.POST("/signup", handler.SignUpCustomer)
		invalidJSON := "{invalid json}"
		req, _ := http.NewRequest(http.MethodPost, "/signup", strings.NewReader(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)

		var respBody map[string]string
		json.Unmarshal(resp.Body.Bytes(), &respBody)
		assert.Equal(t, "Invalid request", respBody["message"])
	})

	t.Run("Invalid Email", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)
		r := gin.Default()
		r.POST("/signup", handler.SignUpCustomer)
		invalidEmailUser := &users.UserRegisterReq{
			Email:    "invalidemail",
			Password: "password123",
			Username: "user",
		}
		reqBody, _ := json.Marshal(invalidEmailUser)
		req, _ := http.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)

		var respBody map[string]string
		json.Unmarshal(resp.Body.Bytes(), &respBody)
		assert.Equal(t, "email pattern is invalid", respBody["message"])
	})

	t.Run("Invalid JSON Format", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)
		r := gin.Default()
		r.POST("/signup-admin", handler.SignUpAdmin)
		invalidJSON := "{invalid json}"
		req, _ := http.NewRequest(http.MethodPost, "/signup-admin", strings.NewReader(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)

		var respBody map[string]string
		err := json.Unmarshal(resp.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, "Bad Request", respBody["error"])
	})

	t.Run("TestSignUpAdminInvalidEmail", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)
		r := gin.Default()
		r.POST("/signup-admin", handler.SignUpAdmin)
		invalidEmailReq := users.UserRegisterReq{
			Email: "invalidemail",
		}
		reqBody, _ := json.Marshal(invalidEmailReq)
		req, _ := http.NewRequest(http.MethodPost, "/signup-admin", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("SignUpAdmin_UsecaseError", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)
		r := gin.Default()
		r.POST("/signup-admin", handler.SignUpAdmin)

		mockError := &users.HTTPError{StatusCode: http.StatusInternalServerError, Message: "Internal Server Error"}
		mockUsecases.On("InsertAdmin", mock.Anything).Return((*users.UserPassport)(nil), mockError).Once()

		validReqBody := `{"email": "admin@example.com", "other_fields": "values"}`
		req, _ := http.NewRequest(http.MethodPost, "/signup-admin", strings.NewReader(validReqBody))

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, mockError.StatusCode, w.Code)
		assert.Contains(t, w.Body.String(), mockError.Message)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)
		r := gin.Default()
		r.POST("/signin", handler.SignIn)

		invalidJSON := "{invalid json}"
		req, _ := http.NewRequest(http.MethodPost, "/signin", strings.NewReader(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)

		var respBody map[string]string
		err := json.Unmarshal(resp.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request", respBody["error"])
	})

	t.Run("Authentication Failure", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)
		mockUsecases.On("GetPassport", mock.Anything).Return((*users.UserPassport)(nil), errors.New("authentication failed"))

		r := gin.Default()
		r.POST("/signin", handler.SignIn)

		reqBody, _ := json.Marshal(users.UserCredential{Email: "user@example.com", Password: "wrongpassword"})
		req, _ := http.NewRequest(http.MethodPost, "/signin", bytes.NewBuffer(reqBody))

		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)

		var respBody map[string]string
		err := json.Unmarshal(resp.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, "Authentication failed", respBody["error"])

		mockUsecases.AssertExpectations(t)
	})

	t.Run("Success Authentication", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)
		someUserId := primitive.NewObjectID()
		mockUsecases.On("GetPassport", validCredentials).Return(&users.UserPassport{
			User: &users.User{
				Id:       someUserId,
				Email:    "pok92deng@example.com",
				Username: "pok92deng",
				Role:     "user",
				RoleId:   1,
			},
			Token: &users.UserToken{
				Id:           someUserId.Hex(),
				AccessToken:  "validAccessToken",
				RefreshToken: "validRefreshToken",
			},
		}, nil)

		r := gin.Default()
		r.POST("/signin", handler.SignIn)

		reqBody, _ := json.Marshal(validCredentials)
		req, _ := http.NewRequest(http.MethodPost, "/signin", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var respBody users.UserPassport
		err := json.Unmarshal(resp.Body.Bytes(), &respBody)
		assert.NoError(t, err)

		mockUsecases.AssertExpectations(t)
	})

	t.Run("Successful Refresh", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)
		mockUsecases.On("RefreshPassport", validRefreshCredential).Return(&users.UserPassport{ /* fill with valid data */ }, nil)

		r := gin.Default()
		r.POST("/refresh", handler.RefreshPassport)

		reqBody, _ := json.Marshal(validRefreshCredential)
		req, _ := http.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var respBody users.UserPassport
		err := json.Unmarshal(resp.Body.Bytes(), &respBody)
		assert.NoError(t, err)

		mockUsecases.AssertExpectations(t)
	})

	t.Run("Invalid JSON Format", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)
		r := gin.Default()
		r.POST("/refresh", handler.RefreshPassport)

		invalidJSON := "{invalid json}"
		req, _ := http.NewRequest(http.MethodPost, "/refresh", strings.NewReader(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Usecase Error", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)
		mockUsecases.On("RefreshPassport", validRefreshCredential).Return(nil, errors.New("some error"))

		r := gin.Default()
		r.POST("/refresh", handler.RefreshPassport)

		reqBody, _ := json.Marshal("Error")
		req, _ := http.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)

	})

	t.Run("Successful UserProfile Retrieval", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)
		userID := primitive.NewObjectID()
		mockUser := &users.User{
			Id:       userID,
			Email:    "user@example.com",
			Username: "testuser",
			Role:     "user",
			RoleId:   1,
		}

		mockUsecases.On("GetUserProfile", userID.Hex()).Return(mockUser, nil)

		r := gin.Default()
		r.GET("/user/:user_id", handler.GerUserProfile)

		req, _ := http.NewRequest(http.MethodGet, "/user/"+userID.Hex(), nil)
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		mockUsecases.AssertCalled(t, "GetUserProfile", userID.Hex())
	})

	t.Run("UserProfile Not Found", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)
		userID := primitive.NewObjectID().Hex()
		err := errors.New("get user failed: mongodb: no rows in result set")

		mockUsecases.On("GetUserProfile", userID).Return((*users.User)(nil), err)

		r := gin.Default()
		r.GET("/user/:user_id", handler.GerUserProfile)

		req, _ := http.NewRequest(http.MethodGet, "/user/"+userID, nil)
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		mockUsecases.AssertCalled(t, "GetUserProfile", userID)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)
		userID := primitive.NewObjectID().Hex()
		err := errors.New("internal server error")

		mockUsecases.On("GetUserProfile", userID).Return((*users.User)(nil), err)

		r := gin.Default()
		r.GET("/user/:user_id", handler.GerUserProfile)

		req, _ := http.NewRequest(http.MethodGet, "/user/"+userID, nil)
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		mockUsecases.AssertCalled(t, "GetUserProfile", userID)
	})

	t.Run("GetAllUserProfile", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)
		mockUsecases.On("GetAllUserProfile").Return([]*users.User{}, nil)

		r := gin.Default()
		r.GET("/users", handler.GetAllUserProfile)

		req, _ := http.NewRequest(http.MethodGet, "/users", nil)
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		mockUsecases.AssertCalled(t, "GetAllUserProfile")
	})

	t.Run("GetAllUserProfile_InternalServerError", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)

		// Set up the expected behavior of the mock
		expectedError := errors.New("internal server error")
		mockUsecases.On("GetAllUserProfile").Return([]*users.User{}, expectedError).Once()

		r := gin.Default()
		r.GET("/user/profiles", handler.GetAllUserProfile)

		req, err := http.NewRequest(http.MethodGet, "/user/profiles", nil)
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected status code to be 500 (Internal Server Error)")
		assert.Contains(t, w.Body.String(), expectedError.Error(), "Expected error message in response")

		mockUsecases.AssertExpectations(t)
	})

	t.Run("CreateRole", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)

		expectedError := errors.New("create role error")
		mockUsecases.On("CreateRole", "1", "TestRole").Return(expectedError).Once()

		r := gin.Default()
		r.POST("/role", handler.CreateRole)

		reqBody := `{"role_id":"1","role":"TestRole"}`
		req, err := http.NewRequest(http.MethodPost, "/role", strings.NewReader(reqBody))
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected internal server error")
		assert.Contains(t, w.Body.String(), expectedError.Error(), "Expected error message in response")

		mockUsecases.AssertExpectations(t)
	})

	t.Run("CreateRole_Success", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)

		mockUsecases.On("CreateRole", "1", "TestRole").Return(nil).Once()

		r := gin.Default()
		r.POST("/role", handler.CreateRole)

		reqBody := `{"role_id":"1","role":"TestRole"}`
		req, err := http.NewRequest(http.MethodPost, "/role", strings.NewReader(reqBody))
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Expected status code to be 201 (Created)")
		assert.Contains(t, w.Body.String(), "Create role successfully", "Expected success message in response")

		mockUsecases.AssertExpectations(t)
	})

	t.Run("CreateRole_InvalidRequest", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)

		r := gin.Default()
		r.POST("/role", handler.CreateRole)

		invalidReqBody := `{"role_id": "1", "role":}`
		req, err := http.NewRequest(http.MethodPost, "/role", strings.NewReader(invalidReqBody))
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code to be 400 (Bad Request)")
		assert.Contains(t, w.Body.String(), "Invalid request format", "Expected error message in response")

	})

	t.Run("UpdateRole_InvalidJSON", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)

		r := gin.Default()
		r.PUT("/role/:user_id", handler.UpdateRole)

		invalidReqBody := `{"role_id": "invalid_json}`
		req, _ := http.NewRequest(http.MethodPut, "/role/1", strings.NewReader(invalidReqBody))

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request format")
	})

	t.Run("UpdateRole_InvalidRoleID", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)

		r := gin.Default()
		r.PUT("/role/:user_id", handler.UpdateRole)

		invalidRoleIdReqBody := `{"role_id": "abc"}`
		req, _ := http.NewRequest(http.MethodPut, "/role/1", strings.NewReader(invalidRoleIdReqBody))

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid role ID")
	})

	t.Run("UpdateRole_UsecaseFailure", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)

		r := gin.Default()
		r.PUT("/role/:user_id", handler.UpdateRole)

		expectedError := errors.New("usecase error")
		mockUsecases.On("UpdateRole", "1", 2, "admin").Return(expectedError).Once()

		validReqBody := `{"role_id": "2", "role": "admin"}`
		req, _ := http.NewRequest(http.MethodPut, "/role/1", strings.NewReader(validReqBody))

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), expectedError.Error())
	})

	t.Run("UpdateRole_Success", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)

		r := gin.Default()
		r.PUT("/role/:user_id", handler.UpdateRole)

		mockUsecases.On("UpdateRole", "1", 2, "admin").Return(nil).Once()

		validReqBody := `{"role_id": "2", "role": "admin"}`
		req, _ := http.NewRequest(http.MethodPut, "/role/1", strings.NewReader(validReqBody))

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Update role successfully")
	})

	t.Run("RefreshPassportAdmin_InvalidJSON", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)

		router := gin.Default()
		router.POST("/refreshPassportAdmin", handler.RefreshPassportAdmin)
		invalidReqBody := `{"invalid_json"`
		req, _ := http.NewRequest(http.MethodPost, "/refreshPassportAdmin", strings.NewReader(invalidReqBody))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request format")
	})

	t.Run("RefreshPassportAdmin_UsecaseError", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)

		router := gin.Default()
		router.POST("/refreshPassportAdmin", handler.RefreshPassportAdmin)

		expectedError := errors.New("usecase error")
		mockUsecases.On("RefreshPassportAdmin", mock.Anything).Return((*users.UserPassport)(nil), expectedError).Once()

		validReqBody := `{"some_field": "some_value"}`
		req, _ := http.NewRequest(http.MethodPost, "/refreshPassportAdmin", strings.NewReader(validReqBody))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Error refreshing passport")
	})

	t.Run("RefreshPassportAdmin_Success", func(t *testing.T) {
		cfg := makeConfig()
		mockUsecases := new(mockUsecases)
		handler := UsersHandler(cfg, mockUsecases)

		router := gin.Default()
		router.POST("/refreshPassportAdmin", handler.RefreshPassportAdmin)

		exampleUser := &users.User{
			Id:       primitive.NewObjectID(),
			Email:    "example@email.com",
			Username: "exampleUser",
			Role:     "user",
			RoleId:   1,
		}
		exampleToken := &users.UserToken{
			Id:           "tokenID",
			AccessToken:  "access-token-string",
			RefreshToken: "refresh-token-string",
		}

		mockPassport := &users.UserPassport{
			User:  exampleUser,
			Token: exampleToken,
		}
		mockUsecases.On("RefreshPassportAdmin", mock.Anything).Return(mockPassport, nil).Once()

		validReqBody := `{"some_field": "some_value"}`
		req, _ := http.NewRequest(http.MethodPost, "/refreshPassportAdmin", strings.NewReader(validReqBody))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

}
