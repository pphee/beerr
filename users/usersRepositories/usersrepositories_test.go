package usersRepositories

import (
	"context"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"pok92deng/config"
	"pok92deng/users"
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

	t.Run("InsertUser", func(t *testing.T) {
		_, err := repo.InsertUser(testUser, false)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("InsertAdmin", func(t *testing.T) {
		_, err := repo.InsertUser(testUser, true)
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

	t.Run("InsertOauth", func(t *testing.T) {
		_, err := repo.InsertUser(testUser, false)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("FindOneOauth", func(t *testing.T) {
		_, err := repo.FindOneUserByEmail(testUser.Email)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("GetProfile", func(t *testing.T) {
		_, err := repo.GetProfile(testUser.Email)
		if err != nil {
			t.Error(err)
		}
	})
}