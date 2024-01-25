package usersRepositories

import (
	"context"
	"fmt"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"pok92deng/config"
	"pok92deng/module/users"
	"strconv"
	"testing"
)

func setupInMemoryMongoDB(t *testing.T) (*mongo.Client, *mongo.Database, func()) {
	mongoServer, err := memongo.Start("4.0.5") // Specify the MongoDB version
	if err != nil {
		t.Fatalf("memongo.Start failed: %s", err)
	}

	opts := options.Client().ApplyURI(mongoServer.URI())
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		mongoServer.Stop()
		t.Fatalf("mongo.Connect failed: %s", err)
	}

	if client == nil {
		mongoServer.Stop()
		t.Fatal("mongo.Client is nil after connect")
	}

	database := client.Database("testdb")

	return client, database, func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Println("Disconnect failed:", err)
		}
		mongoServer.Stop()
	}
}

func makeConfig() config.IConfig {
	cfg := config.LoadConfig("/Users/p/Goland/pok92deng/.env")
	return cfg
}

func TestNewMongoDB(t *testing.T) {
	_, database, cleanup := setupInMemoryMongoDB(t)
	defer cleanup()

	cfg := makeConfig()

	repo := UsersRepository(cfg, database)

	StoreMongo(t, repo)
}

func StoreMongo(t *testing.T, repo UserRepository) {
	testUser := &users.UserRegisterReq{
		Email:    "phee@gmail.com",
		Password: "123456",
		Username: "phee",
	}

	testAdmin := &users.UserRegisterReq{
		Email:    "p@gmail.com",
		Password: "123456",
		Username: "p",
	}

	userID := primitive.NewObjectID()

	oauthUser := &users.UserPassport{
		User: &users.User{
			Id:       userID,
			Email:    "pheeDan@gmail.com",
			Username: "testuser",
			Role:     "user",
			RoleId:   1,
		},
		Token: &users.UserToken{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
		},
	}

	var validUserID string

	newRoleId := 4

	t.Run("InsertUser", func(t *testing.T) {
		user, err := repo.InsertUser(testUser, false)
		if err != nil {
			t.Error(err)
		} else {
			validUserID = user.User.Id.Hex()
		}
	})

	t.Run("checkUserExistenceForDuplicateInsert", func(t *testing.T) {
		_, err := repo.InsertUser(testUser, false)
		if err == nil {
			t.Error("Expected error for existing user, got nil")
		}
	})

	t.Run("checkUserExistenceForDuplicateInsert", func(t *testing.T) {
		err := repo.CheckUserExistence(testUser.Email, testUser.Username)
		if err == nil {
			t.Error("Expected error for existing user, got nil")
		}
	})

	t.Run("ErrorInAdminOrCustomer", func(t *testing.T) {
		_, err := repo.InsertUser(testUser, true)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("ErrorInResult", func(t *testing.T) {
		_, err := repo.InsertUser(testUser, false)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("FindOneUserByEmail", func(t *testing.T) {
		_, err := repo.FindOneUserByEmail(testUser.Email)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("FindOneUserByUsername", func(t *testing.T) {
		_, err := repo.FindOneUserByUsername(testUser.Username)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("InsertAdmin", func(t *testing.T) {
		_, err := repo.InsertUser(testAdmin, true)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("InsertOauth", func(t *testing.T) {
		err := repo.InsertOauth(oauthUser)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("UpdateOauth", func(t *testing.T) {
		oauth, err := repo.FindOneOauth("refresh-token")

		oauthUserUpdate := &users.UserToken{
			Id:           oauth.Id, // Correctly referencing the ID
			AccessToken:  "access-tokenUpdate",
			RefreshToken: "refresh-tokenUpdate",
		}

		err = repo.UpdateOauth(oauthUserUpdate)
		if err != nil {
			t.Error("UpdateOauth failed:", err)
		}
	})

	t.Run("FindOneOauth", func(t *testing.T) {
		_, err := repo.FindOneOauth("refresh-tokenUpdate")
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("FindOneUserByEmail", func(t *testing.T) {
		_, err := repo.FindOneUserByEmail(testUser.Email)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("GetProfile", func(t *testing.T) {
		if validUserID == "" {
			t.Skip("Valid user ID not available")
		} else {
			_, err := repo.GetProfile(validUserID)
			if err != nil {
				t.Error(err)
			}
		}
	})

	t.Run("GetAllUserProfile", func(t *testing.T) {
		profiles, err := repo.GetAllUserProfile()
		if err != nil {
			t.Errorf("Error fetching profiles: %v", err)
		}

		for _, profile := range profiles {
			fmt.Println(profile)
		}
		if len(profiles) != 2 {
			t.Errorf("Expected 2 profiles, got %d", len(profiles))
		}
	})

	t.Run("UpdateRole", func(t *testing.T) {
		// Fetch all user profiles
		profiles, err := repo.GetAllUserProfile()
		if err != nil {
			t.Errorf("Error fetching profiles: %v", err)
			return
		}

		var profileIDs []string

		for _, profile := range profiles {
			profileIDs = append(profileIDs, profile.Id.Hex())
		}

		err = repo.UpdateRole(profileIDs[0], newRoleId, "manager")
		if err != nil {
			t.Errorf("Failed to update user role: %v", err)
			return
		}

		updatedUser, err := repo.GetProfile(profileIDs[0])
		if err != nil {
			t.Errorf("Failed to retrieve updated user: %v", err)
			return
		}

		if updatedUser != nil && updatedUser.RoleId != newRoleId {
			t.Errorf("Role ID was not updated correctly. Expected %d, got %d", newRoleId, updatedUser.RoleId)
		}
	})

	// Step 5: Test error handling
	t.Run("UpdateRoleWithInvalidID", func(t *testing.T) {
		err := repo.UpdateRole("invalidID", newRoleId, "manager")
		if err == nil {
			t.Errorf("Expected error with invalid user ID, but got none")
		}
	})

	t.Run("UpdateRoleWithNonExistingID", func(t *testing.T) {
		nonExistingID := "5f50c31f5b5f5b5f5b5f5b5f" // Example non-existing ObjectID
		err := repo.UpdateRole(nonExistingID, newRoleId, "manager")
		if err == nil {
			t.Errorf("Expected error with non-existing user ID, but got none")
		}
	})

	t.Run("CreateRole", func(t *testing.T) {
		err := repo.CreateRole(strconv.Itoa(newRoleId), "testRole")
		if err != nil {
			t.Error(err)
		}
	})

}
