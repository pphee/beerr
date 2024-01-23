package middlewaresRepositories

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"pok92deng/config"
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

	repo := MiddlewaresRepository(cfg, database)

	StoreMongo(t, repo, database, cfg)
}

func StoreMongo(t *testing.T, repo IMiddlewaresRepository, database *mongo.Database, cfg config.IConfig) {
	ctx := context.Background()
	collectionName := cfg.Db().RolesCollection()
	t.Run("FindAccessToken", func(t *testing.T) {
		userId := primitive.NewObjectID() // Use ObjectID directly
		accessToken := "someAccessToken"

		collectionName := cfg.Db().SigninsCollection()
		_, err := database.Collection(collectionName).InsertOne(context.Background(), bson.M{
			"user_id":      userId,
			"access_token": accessToken,
		})
		require.NoError(t, err)

		result := repo.FindAccessToken(userId.Hex(), accessToken) // Convert ObjectID to Hex string
		expectedValue := true
		assert.Equal(t, expectedValue, result)
	})

	t.Run("InvalidUserId", func(t *testing.T) {
		invalidUserId := "invalidUserId"
		accessToken := "someAccessToken"

		result := repo.FindAccessToken(invalidUserId, accessToken)

		assert.False(t, result, "Expected false when userId cannot be converted to ObjectId")
	})

	t.Run("ValidRoleId", func(t *testing.T) {

		_, err := database.Collection(collectionName).InsertOne(ctx, bson.M{
			"roleId": "1",
			"role":   "TestRole",
		})
		require.NoError(t, err)

		roles, err := repo.FindRole(ctx, "1")
		require.NoError(t, err)
		require.Len(t, roles, 1)
	})

	t.Run("InvalidRoleId", func(t *testing.T) {
		invalidRoleId := "nonExistingRole"

		roles, err := repo.FindRole(ctx, invalidRoleId)
		require.NoError(t, err)
		assert.Empty(t, roles)
	})

}
