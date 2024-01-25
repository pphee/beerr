package usersUsecases

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"pok92deng/config"
	"pok92deng/module/users"
	auth "pok92deng/pkg"
	"strings"
	"testing"
)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) UpdateRole(userId string, roleId int, role string) error {
	args := m.Called(userId, roleId, role)
	return args.Error(0)
}

func (m *mockRepository) CreateRole(roleId, role string) error {
	args := m.Called(roleId, role)
	return args.Error(0)
}

func (m *mockRepository) GetAllUserProfile() ([]*users.User, error) {
	args := m.Called()
	return args.Get(0).([]*users.User), args.Error(1)
}

func (m *mockRepository) InsertUser(req *users.UserRegisterReq, isAdmin bool) (*users.UserPassport, error) {
	args := m.Called(req, isAdmin)
	return args.Get(0).(*users.UserPassport), args.Error(1)
}

func (m *mockRepository) FindOneUserByUsername(username string) (*users.UserCredentialCheck, error) {
	args := m.Called(username)
	return args.Get(0).(*users.UserCredentialCheck), args.Error(1)
}

func (m *mockRepository) FindOneUserByEmail(email string) (*users.UserCredentialCheck, error) {
	args := m.Called(email)
	return args.Get(0).(*users.UserCredentialCheck), args.Error(1)
}

func (m *mockRepository) InsertOauth(req *users.UserPassport) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *mockRepository) FindOneOauth(refreshToken string) (*users.Oauth, error) {
	args := m.Called(refreshToken)
	return args.Get(0).(*users.Oauth), args.Error(1)
}

func (m *mockRepository) GetProfile(userId string) (*users.User, error) {
	args := m.Called(userId)
	return args.Get(0).(*users.User), args.Error(1)
}

func (m *mockRepository) UpdateOauth(req *users.UserToken) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *mockRepository) CheckUserExistence(email, username string) error {
	args := m.Called(email, username)
	return args.Error(0)

}

func (m *mockRepository) NewAuth(cfg config.IConfig, userClaims *users.UserClaims, tokenType auth.TokenType) (string, error) {
	args := m.Called(cfg, userClaims, tokenType)
	return args.String(0), args.Error(1)
}

func makeConfig() config.IConfig {
	cfg := config.LoadConfig("/Users/p/Goland/pok92deng/.env")
	return cfg
}

func TestUsersUsecases(t *testing.T) {
	testUser := &users.UserRegisterReq{
		Email:    "phee@gmail.com",
		Password: "phee007",
		Username: "phee",
	}

	testAdmin := &users.UserRegisterReq{
		Email:    "Admin@gmail.com",
		Password: "Admin",
		Username: "Admin",
	}

	t.Run("InsertCustomer", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)
		mockRepo.On("InsertUser", testUser, false).Return(&users.UserPassport{}, nil)
		_, err := usersUsecase.InsertCustomer(testUser)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("InsertCustomerError", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)
		mockRepo.On("InsertUser", testUser, false).Return((*users.UserPassport)(nil), errors.New("mock error"))
		_, err := usersUsecase.InsertCustomer(testUser)
		assert.Error(t, err)
	})

	t.Run("InsertAdmin", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)
		mockRepo.On("InsertUser", testAdmin, true).Return(&users.UserPassport{}, nil)
		_, err := usersUsecase.InsertAdmin(testAdmin)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("InsertAdminError", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)
		mockRepo.On("InsertUser", testAdmin, true).Return((*users.UserPassport)(nil), errors.New("mock error"))
		_, err := usersUsecase.InsertAdmin(testAdmin)
		assert.Error(t, err)
	})

	t.Run("GetPassport", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testUser.Password), bcrypt.DefaultCost)
		if err != nil {
			t.Fatalf("Failed to hash password: %v", err)
		}

		mockUser := &users.UserCredentialCheck{
			Email:    testUser.Email,
			Username: testUser.Username,
			Password: string(hashedPassword),
			Role:     "user",
			RoleId:   1,
		}

		hashedAdminPassword, err := bcrypt.GenerateFromPassword([]byte(testAdmin.Password), bcrypt.DefaultCost)
		if err != nil {
			t.Fatalf("Failed to hash admin password: %v", err)
		}

		mockAdmin := &users.UserCredentialCheck{
			Email:    testAdmin.Email,
			Username: testAdmin.Username,
			Password: string(hashedAdminPassword),
			Role:     "admin",
			RoleId:   2,
		}

		mockRepo.On("FindOneUserByEmail", testUser.Email).Return(mockUser, nil)
		mockRepo.On("FindOneUserByEmail", testAdmin.Email).Return(mockAdmin, nil)
		mockRepo.On("InsertOauth", mock.AnythingOfType("*users.UserPassport")).Return(nil)

		passport, err := usersUsecase.GetPassport(&users.UserCredential{Email: testUser.Email, Password: testUser.Password})
		if err != nil {
			t.Errorf("GetPassport failed for regular user: %v", err)
		} else if passport == nil || passport.User == nil {
			t.Error("GetPassport returned a nil passport or nil User for regular user")
		} else if passport.User.Email != testUser.Email {
			t.Errorf("Expected email %v for regular user, got %v", testUser.Email, passport.User.Email)
		}

		adminPassport, err := usersUsecase.GetPassport(&users.UserCredential{Email: testAdmin.Email, Password: testAdmin.Password})
		if err != nil {
			t.Errorf("GetPassport failed for admin user: %v", err)
		} else if adminPassport == nil || adminPassport.User == nil {
			t.Error("GetPassport returned a nil passport or nil User for admin user")
		} else if adminPassport.User.Email != testAdmin.Email {
			t.Errorf("Expected email %v for admin user, got %v", testAdmin.Email, adminPassport.User.Email)
		}

		mockRepo.AssertExpectations(t)
	})

	t.Run("GetPassportError", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)

		mockRepo.On("FindOneUserByEmail", testUser.Email).Return((*users.UserCredentialCheck)(nil), errors.New("mock error"))

		_, err := usersUsecase.GetPassport(&users.UserCredential{Email: testUser.Email, Password: testUser.Password})
		assert.Error(t, err)
	})

	t.Run("GetPassportErrorInsertOauthError", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testUser.Password), bcrypt.DefaultCost)
		if err != nil {
			t.Fatalf("Failed to hash password: %v", err)
		}

		mockUser := &users.UserCredentialCheck{
			Email:    testUser.Email,
			Username: testUser.Username,
			Password: string(hashedPassword),
			Role:     "user",
			RoleId:   1,
		}

		mockRepo.On("FindOneUserByEmail", testUser.Email).Return(mockUser, nil)
		mockRepo.On("InsertOauth", mock.AnythingOfType("*users.UserPassport")).Return(errors.New("mock error"))

		_, err = usersUsecase.GetPassport(&users.UserCredential{Email: testUser.Email, Password: testUser.Password})
		assert.Error(t, err)
	})

	t.Run("ErrCompareHashAndPassword", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testUser.Password), bcrypt.DefaultCost)
		if err != nil {
			t.Fatalf("Failed to hash password: %v", err)
		}

		mockUser := &users.UserCredentialCheck{
			Email:    testUser.Email,
			Username: testUser.Username,
			Password: string(hashedPassword),
			Role:     "user",
			RoleId:   1,
		}

		mockRepo.On("FindOneUserByEmail", testUser.Email).Return(mockUser, nil)

		_, err = usersUsecase.GetPassport(&users.UserCredential{Email: testUser.Email, Password: "wrongPassword"})
		if err == nil {
			t.Errorf("Expected an error for wrong password, got none")
		}

		assert.Equal(t, "password is invalid", err.Error())
	})

	t.Run("RefreshPassport", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)
		someUserId := primitive.NewObjectID()

		userClaims := &users.UserClaims{
			Id:     someUserId,
			Role:   "user",
			RoleId: 1,
		}

		tokenGenerator, _ := auth.NewAuth(
			auth.Refresh,
			cfg.Jwt(),
			userClaims,
		)
		validRefreshToken := tokenGenerator.SignToken()

		mockOauth := &users.Oauth{
			UserId: someUserId.Hex(),
		}

		mockProfile := &users.User{
			Id:       someUserId,
			Email:    "phee@gmail.com",
			Username: "phee",
			Role:     "user",
			RoleId:   1,
		}

		mockRepo.On("FindOneOauth", validRefreshToken).Return(mockOauth, nil)
		mockRepo.On("GetProfile", someUserId.Hex()).Return(mockProfile, nil)
		mockRepo.On("UpdateOauth", mock.AnythingOfType("*users.UserToken")).Return(nil)

		_, err := usersUsecase.RefreshPassport(&users.UserRefreshCredential{RefreshToken: validRefreshToken})
		if err != nil {
			t.Errorf("RefreshPassport failed: %v", err)
		}

		mockRepo.AssertExpectations(t)
	})

	t.Run("RefreshPassportFindOneOauthError", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)

		tokenGenerator, _ := auth.NewAuth(
			auth.Refresh,
			cfg.Jwt(),
			&users.UserClaims{
				Id:     primitive.NewObjectID(),
				RoleId: 1,
			},
		)
		invalidRefreshToken := tokenGenerator.SignToken()

		mockRepo.On("FindOneOauth", invalidRefreshToken).Return((*users.Oauth)(nil), errors.New("mock error"))

		_, err := usersUsecase.RefreshPassport(&users.UserRefreshCredential{RefreshToken: invalidRefreshToken})
		assert.Error(t, err)
	})

	t.Run("RefreshPassportGetProfileError", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)

		tokenGenerator, _ := auth.NewAuth(
			auth.Refresh,
			cfg.Jwt(),
			&users.UserClaims{
				Id:     primitive.NewObjectID(),
				Role:   "user",
				RoleId: 1,
			},
		)
		validRefreshToken := tokenGenerator.SignToken()

		mockOauth := &users.Oauth{
			UserId: primitive.NewObjectID().Hex(),
		}

		mockRepo.On("FindOneOauth", validRefreshToken).Return(mockOauth, nil)
		mockRepo.On("GetProfile", mockOauth.UserId).Return((*users.User)(nil), errors.New("mock error"))

		_, err := usersUsecase.RefreshPassport(&users.UserRefreshCredential{RefreshToken: validRefreshToken})
		assert.Error(t, err)
	})

	t.Run("RefreshPassportUpdateOauthError", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)

		tokenGenerator, _ := auth.NewAuth(
			auth.Refresh,
			cfg.Jwt(),
			&users.UserClaims{
				Id:     primitive.NewObjectID(),
				Role:   "user",
				RoleId: 1,
			},
		)
		validRefreshToken := tokenGenerator.SignToken()

		mockOauth := &users.Oauth{
			UserId: primitive.NewObjectID().Hex(),
		}

		mockProfile := &users.User{
			Id:       primitive.NewObjectID(),
			Email:    "phee@gmail.com",
			Username: "phee",
			Role:     "admin",
			RoleId:   2,
		}

		mockRepo.On("FindOneOauth", validRefreshToken).Return(mockOauth, nil)
		mockRepo.On("GetProfile", mockOauth.UserId).Return(mockProfile, nil)

		mockRepo.On("UpdateOauth", mock.AnythingOfType("*users.UserToken")).Return(errors.New("mock error"))

		_, err := usersUsecase.RefreshPassport(&users.UserRefreshCredential{RefreshToken: validRefreshToken})
		assert.Error(t, err)
	})

	t.Run("RefreshPassportAdmin", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)
		someUserId := primitive.NewObjectID()

		userClaims := &users.UserClaims{
			Id:     someUserId,
			Role:   "admin",
			RoleId: 2,
		}

		tokenGenerator, _ := auth.NewAuth(
			auth.RefreshTokenAdmin,
			cfg.Jwt(),
			userClaims,
		)
		validRefreshToken := tokenGenerator.SignToken()

		mockOauth := &users.Oauth{
			UserId: someUserId.Hex(),
		}

		mockProfile := &users.User{
			Id:       someUserId,
			Email:    "phee@gmail.com",
			Username: "phee",
			Role:     "admin",
			RoleId:   2,
		}

		mockRepo.On("FindOneOauth", validRefreshToken).Return(mockOauth, nil)
		mockRepo.On("GetProfile", someUserId.Hex()).Return(mockProfile, nil)

		mockRepo.On("UpdateOauth", mock.AnythingOfType("*users.UserToken")).Return(nil)

		_, err := usersUsecase.RefreshPassportAdmin(&users.UserRefreshCredential{RefreshToken: validRefreshToken})
		if err != nil {
			t.Errorf("RefreshPassport failed: %v", err)
		}
		mockRepo.AssertExpectations(t)

	})

	t.Run("RefreshPassportAdminFindOneOauthError", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)

		tokenGenerator, _ := auth.NewAuth(
			auth.RefreshTokenAdmin,
			cfg.Jwt(),
			&users.UserClaims{
				Id:     primitive.NewObjectID(),
				Role:   "admin",
				RoleId: 2,
			},
		)
		invalidRefreshToken := tokenGenerator.SignToken()

		mockRepo.On("FindOneOauth", invalidRefreshToken).Return((*users.Oauth)(nil), errors.New("mock error"))

		_, err := usersUsecase.RefreshPassportAdmin(&users.UserRefreshCredential{RefreshToken: invalidRefreshToken})
		assert.Error(t, err)
	})

	t.Run("RefreshPassportAdminGetProfileError", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)

		tokenGenerator, _ := auth.NewAuth(
			auth.RefreshTokenAdmin,
			cfg.Jwt(),
			&users.UserClaims{
				Id:     primitive.NewObjectID(),
				Role:   "admin",
				RoleId: 2,
			},
		)
		validRefreshToken := tokenGenerator.SignToken()

		mockOauth := &users.Oauth{
			UserId: primitive.NewObjectID().Hex(),
		}

		mockRepo.On("FindOneOauth", validRefreshToken).Return(mockOauth, nil)
		mockRepo.On("GetProfile", mockOauth.UserId).Return((*users.User)(nil), errors.New("mock error"))

		_, err := usersUsecase.RefreshPassportAdmin(&users.UserRefreshCredential{RefreshToken: validRefreshToken})
		assert.Error(t, err)
	})

	t.Run("RefreshPassportAdminUpdateOauthError", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)

		tokenGenerator, _ := auth.NewAuth(
			auth.RefreshTokenAdmin,
			cfg.Jwt(),
			&users.UserClaims{
				Id:     primitive.NewObjectID(),
				RoleId: 2,
			},
		)
		validRefreshToken := tokenGenerator.SignToken()

		mockOauth := &users.Oauth{
			UserId: primitive.NewObjectID().Hex(),
		}

		mockProfile := &users.User{
			Id:       primitive.NewObjectID(),
			Email:    "phee@gmail.com",
			Username: "phee",
			Role:     "admin",
			RoleId:   2,
		}

		mockRepo.On("FindOneOauth", validRefreshToken).Return(mockOauth, nil)
		mockRepo.On("GetProfile", mockOauth.UserId).Return(mockProfile, nil)

		mockRepo.On("UpdateOauth", mock.AnythingOfType("*users.UserToken")).Return(errors.New("mock error"))

		_, err := usersUsecase.RefreshPassportAdmin(&users.UserRefreshCredential{RefreshToken: validRefreshToken})
		assert.Error(t, err)
	})

	t.Run("GetUserProfile", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)
		mockRepo.On("GetProfile", testUser.Email).Return(&users.User{}, nil)
		_, err := usersUsecase.GetUserProfile(testUser.Email)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("GetUserProfileError", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)
		mockRepo.On("GetProfile", testUser.Email).Return((*users.User)(nil), errors.New("mock error"))
		profile, err := usersUsecase.GetUserProfile(testUser.Email)
		assert.Error(t, err)

		assert.Nil(t, profile)
	})

	t.Run("InsertCustomerPasswordError", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)
		testPasswordError := &users.UserRegisterReq{
			Email:    "phee@gmail.com",
			Password: strings.Repeat("a", 73),
			Username: "phee",
		}

		mockRepo.On("InsertUser", testPasswordError, false).Return(&users.UserPassport{}, nil)

		_, err := usersUsecase.InsertCustomer(testPasswordError)

		if err == nil {
			t.Errorf("Expected an error for password length exceeding 72 bytes, got none")
		}

		assert.Equal(t, "hashed password failed: bcrypt: password length exceeds 72 bytes", err.Error())
	})

	t.Run("InsertAdminPasswordError", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)
		testPasswordError := &users.UserRegisterReq{
			Email:    "phee@gmail.com",
			Password: strings.Repeat("a", 73),
			Username: "phee",
		}

		mockRepo.On("InsertUser", testPasswordError, true).Return(&users.UserPassport{}, nil)

		_, err := usersUsecase.InsertAdmin(testPasswordError)

		if err == nil {
			t.Errorf("Expected an error for password length exceeding 72 bytes, got none")
		}
		assert.Equal(t, "hashed password failed: bcrypt: password length exceeds 72 bytes", err.Error())
	})

	t.Run("GetAllUserProfile", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)

		mockRepo.On("GetAllUserProfile").Return([]*users.User{}, nil)

		_, err := usersUsecase.GetAllUserProfile()
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("GetAllUserProfileError", func(t *testing.T) {
		t.Run("GetAllUserProfileSuccess", func(t *testing.T) {
			// Setting up the configuration and mock repository
			cfg := makeConfig()
			mockRepo := new(mockRepository)
			usersUsecase := UsersUsecase(cfg, mockRepo)

			mockRepo.On("GetAllUserProfile").Return([]*users.User{}, nil)

			result, err := usersUsecase.GetAllUserProfile()

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Len(t, result, 0)

			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("UpdateRole", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)

		mockRepo.On("UpdateRole", testUser.Email, 1, "user").Return(nil)

		err := usersUsecase.UpdateRole(testUser.Email, 1, "user")
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("CreateRole", func(t *testing.T) {
		cfg := makeConfig()
		mockRepo := new(mockRepository)
		usersUsecase := UsersUsecase(cfg, mockRepo)

		mockRepo.On("CreateRole", "1", "admin").Return(nil)

		err := usersUsecase.CreateRole("1", "admin")
		if err != nil {
			t.Error(err)
		}
	})

}
