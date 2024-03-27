package usersHandlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/http/middleware"
	"golang.org/x/exp/slog"
	"io/ioutil"
	"log"
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
	CreateUserProfile(c *gin.Context)
	DeleteUserZitadel(c *gin.Context)
	GetUserZitadel(c *gin.Context)
	ImportUserToZitadel(c *gin.Context)
	LockUserInZitadel(c *gin.Context)
	UnlockUserInZitadel(c *gin.Context)
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

		userProfile := fetchUserProfile(userID)

		c.JSON(http.StatusOK, gin.H{
			"response": userProfile,
			"message":  "get user profile successfully",
		})

	})(c)
}

func fetchUserProfile(userID string) interface{} {
	url := fmt.Sprintf("https://auth.cloudsoft.co.th/management/v1/users/%s", userID)

	method := "GET"

	fmt.Println("url", url)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return gin.H{"error": "Failed to create request"}
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer klSUdCOoVNgrjIHGmGh7H0psoAvdMUSyEXdqqqJ5EA9lyIq79LHh7YFkDEjqo4o00uqBbrk") // Replace <TOKEN> with your actual token

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return gin.H{"error": "Failed to execute request"}
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return gin.H{"error": "Failed to read response body"}
	}

	fmt.Println(string(body))

	return string(body)
}

func (h *usersHandler) GetUserZitadel(c *gin.Context) {
	userID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	url := fmt.Sprintf("https://auth.cloudsoft.co.th/management/v1/users/%s", userID)

	method := "GET"

	fmt.Println("url", url)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer klSUdCOoVNgrjIHGmGh7H0psoAvdMUSyEXdqqqJ5EA9lyIq79LHh7YFkDEjqo4o00uqBbrk") // Replace with your actual token

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute request"})
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response": string(body),
		"message":  "get user profile successfully",
	})
}

func (h *usersHandler) CreateUserProfile(c *gin.Context) {
	var userProfile users.UserProfile
	if err := c.ShouldBindJSON(&userProfile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url := "https://auth.cloudsoft.co.th/management/v1/users/human/_import"
	method := "POST"

	payloadBytes, err := json.Marshal(userProfile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error marshalling JSON"})
		return
	}
	payload := bytes.NewReader(payloadBytes)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer klSUdCOoVNgrjIHGmGh7H0psoAvdMUSyEXdqqqJ5EA9lyIq79LHh7YFkDEjqo4o00uqBbrk") // Use the actual token

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute request"})
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	var responseBody map[string]interface{}
	if err := json.Unmarshal(body, &responseBody); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing response JSON"})
		return
	}

	if code, exists := responseBody["code"].(float64); exists && code == 6 {
		c.JSON(http.StatusConflict, gin.H{
			"error":   responseBody["message"],
			"details": responseBody["details"],
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"response": responseBody,
			"message":  "User profile created successfully",
		})
	}
}

func (h *usersHandler) DeleteUserZitadel(c *gin.Context) {
	userID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	url := fmt.Sprintf("https://auth.cloudsoft.co.th/management/v1/users/%s", userID)
	method := "DELETE"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer klSUdCOoVNgrjIHGmGh7H0psoAvdMUSyEXdqqqJ5EA9lyIq79LHh7YFkDEjqo4o00uqBbrk") // Replace with your actual token

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute request"})
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response": string(body),
		"message":  "User deleted successfully",
	})
}

func (h *usersHandler) ImportUserToZitadel(c *gin.Context) {
	users, err := h.usersUsecase.ImportUsersFromMongo()
	if err != nil {
		log.Printf("Error fetching users from MongoDB: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users from MongoDB"})
		return
	}

	for _, user := range users {
		h.importUserToZitadel(c, user)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Users imported successfully to ZITADEL"})
}

func (h *usersHandler) importUserToZitadel(c *gin.Context, user users.UserZitadel) {
	url := "https://auth.cloudsoft.co.th/management/v1/users/human/_import"

	jsonPayload := fmt.Sprintf(`{
        "userName": "%s",
        "hashedPassword": {
            "value": "%s"
        },
        "profile": {
            "firstName": "FirstNameExample",
            "lastName": "LastNameExample"
        },
        "email": {
            "email": "%s",
            "isEmailVerified": false
        }
    }`, user.Username, user.HashedPassword, user.Email)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, strings.NewReader(jsonPayload))
	if err != nil {
		log.Printf("Error occurred while creating HTTP request: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer klSUdCOoVNgrjIHGmGh7H0psoAvdMUSyEXdqqqJ5EA9lyIq79LHh7YFkDEjqo4o00uqBbrk") // Replace with actual access token

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error occurred while sending request to ZITADEL: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request"})
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error occurred while reading response body: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("ZITADEL responded with status: %d, body: %s", resp.StatusCode, string(body))
		c.JSON(resp.StatusCode, gin.H{"error": "Failed to import user", "response": string(body)})
		return
	}

	fmt.Println("Response from ZITADEL:", string(body))
}

func (h *usersHandler) LockUserInZitadel(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	url := fmt.Sprintf("https://%s/management/v1/users/%s/_lock", "auth.cloudsoft.co.th", userID)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, strings.NewReader(`{}`))
	if err != nil {
		fmt.Println("Error creating request:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating request"})
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer klSUdCOoVNgrjIHGmGh7H0psoAvdMUSyEXdqqqJ5EA9lyIq79LHh7YFkDEjqo4o00uqBbrk")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending request"})
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading response body"})
		return
	}

	var responseBody map[string]interface{}
	if err := json.Unmarshal(body, &responseBody); err != nil {
		fmt.Println("Error parsing response body:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing response body"})
		return
	}

	if code, exists := responseBody["code"].(float64); exists && code == 6 {
		c.JSON(http.StatusConflict, gin.H{
			"error":   responseBody["message"],
			"details": responseBody["details"],
		})
		return
	}

	fmt.Println("Response from ZITADEL:", string(body))
	c.JSON(http.StatusOK, gin.H{
		"message":  "User locked successfully",
		"response": responseBody,
	})
}

func (h *usersHandler) UnlockUserInZitadel(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	url := fmt.Sprintf("https://%s/management/v1/users/%s/_unlock", "auth.cloudsoft.co.th", userID)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, strings.NewReader(`{}`))
	if err != nil {
		fmt.Println("Error creating request:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating request"})
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer klSUdCOoVNgrjIHGmGh7H0psoAvdMUSyEXdqqqJ5EA9lyIq79LHh7YFkDEjqo4o00uqBbrk")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending request"})
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading response body"})
		return
	}

	var responseBody map[string]interface{}
	if err := json.Unmarshal(body, &responseBody); err != nil {
		fmt.Println("Error unmarshalling response body:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing response body"})
		return
	}

	if code, exists := responseBody["code"].(float64); exists && code == 6 {
		c.JSON(http.StatusConflict, gin.H{
			"error":   responseBody["message"],
			"details": responseBody["details"],
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"response": responseBody,
			"message":  "User unlocked successfully",
		})
	}
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
