package repository

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"mime/multipart"
	"pok92deng/config"
	"pok92deng/module/product"
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

	repo := NewproductsRepository(cfg, database)

	StoreMongo(t, repo)
}

func StoreMongo(t *testing.T, repo BeerRepository) {
	fileHeader := &multipart.FileHeader{
		Filename: "test_beer_image.jpg",
		Header:   make(map[string][]string),
	}

	testBeer := model.Beer{
		ID:        primitive.NewObjectID(),
		Name:      "chang",
		Category:  "IPA",
		Detail:    "A test beer with citrus and pine notes",
		Image:     fileHeader,
		ImagePath: "test.jpg",
	}

	t.Run("InsertBeer for testing", func(t *testing.T) {
		id, err := repo.InsertBeer(context.Background(), testBeer)
		require.NoError(t, err)
		assert.NotEqual(t, primitive.NilObjectID, id)
		testBeer.ID = id
	})

	t.Run("InsertBeer with BeerNameExists error", func(t *testing.T) {
		duplicateBeer := model.Beer{
			Name:      testBeer.Name,
			Category:  testBeer.Category,
			Detail:    testBeer.Detail,
			Image:     testBeer.Image,
			ImagePath: testBeer.ImagePath,
		}
		_, err := repo.InsertBeer(context.Background(), duplicateBeer)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "a beer with this name already exists")
	})

	t.Run("InsertBeer with empty name", func(t *testing.T) {
		_, err := repo.InsertBeer(context.Background(), model.Beer{})
		require.Error(t, err, "name must not be empty")
	})

	t.Run("InsertBeer with BeerNameExists error", func(t *testing.T) {
		_, err := repo.InsertBeer(context.Background(), testBeer)
		require.Error(t, err)
	})

	t.Run("InsertBeer with InsertOne error", func(t *testing.T) {
		_, err := repo.InsertBeer(context.Background(), testBeer)
		require.Error(t, err)
	})

	t.Run("FindBeer", func(t *testing.T) {
		products, err := repo.FindBeer(context.Background(), testBeer.ID)
		require.NoError(t, err)
		assert.Equal(t, testBeer.ID, products.ID)
		assert.Equal(t, testBeer.ID, products.ID)
		assert.Equal(t, testBeer.Name, products.Name)
		assert.Equal(t, testBeer.Category, products.Category)
		assert.Equal(t, testBeer.Detail, products.Detail)
		assert.Equal(t, testBeer.ImagePath, products.ImagePath)
		assert.Equal(t, testBeer.Image, products.Image)
	})

	t.Run("FindBeer with error", func(t *testing.T) {
		_, err := repo.FindBeer(context.Background(), primitive.NilObjectID)
		require.Error(t, err)
	})

	t.Run("UpdateBeer", func(t *testing.T) {
		fileHeaderUpdate := &multipart.FileHeader{
			Filename: "test_beer_image_update.jpg",
			Header:   make(map[string][]string),
		}
		updatedBeer := model.Beer{
			ID:       testBeer.ID,
			Name:     "leo",
			Category: "IPA",
			Detail:   "Updated details",
			Image:    fileHeaderUpdate,
		}
		err := repo.UpdateBeer(context.Background(), testBeer.ID, updatedBeer)
		require.NoError(t, err)
	})

	t.Run("UpdateBeer with image nil", func(t *testing.T) {
		updatedBeer := model.Beer{
			ID:       testBeer.ID,
			Name:     "leo",
			Category: "IPA",
			Detail:   "Updated details",
			Image:    nil,
		}
		err := repo.UpdateBeer(context.Background(), testBeer.ID, updatedBeer)
		require.NoError(t, err)
	})

	t.Run("UpdateBeer with error", func(t *testing.T) {
		err := repo.UpdateBeer(context.Background(), primitive.NilObjectID, testBeer)
		require.Error(t, err)
	})

	t.Run("BeerNameExists", func(t *testing.T) {
		exists, err := repo.BeerNameExists(context.Background(), "leo")
		fmt.Println(exists)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("BeerNameExists with empty name", func(t *testing.T) {
		_, err := repo.BeerNameExists(context.Background(), "")
		require.Error(t, err, "name must not be empty")
	})

	t.Run("BeerNameExists with error", func(t *testing.T) {
		_, err := repo.BeerNameExists(context.Background(), "")
		require.Error(t, err)
	})

	t.Run("DeleteBeer", func(t *testing.T) {
		err := repo.DeleteBeer(context.Background(), testBeer.ID)
		require.NoError(t, err)
	})

	t.Run("DeleteBeer with empty ID", func(t *testing.T) {
		err := repo.DeleteBeer(context.Background(), primitive.NilObjectID)
		require.Error(t, err, "id must not be empty")
	})

	t.Run("DeleteBeer with error", func(t *testing.T) {
		err := repo.DeleteBeer(context.Background(), primitive.NewObjectID())
		require.Error(t, err, "test")
	})

	t.Run("FilterAndPaginateBeers", func(t *testing.T) {
		ctx := context.Background()
		name := "Test Beer"
		page, limit := int64(1), int64(5)

		for i := 0; i < 5; i++ {
			testBeer := model.Beer{
				Name: fmt.Sprintf("Test Beer %d", i),
			}
			_, err := repo.InsertBeer(ctx, testBeer)
			require.NoError(t, err)
		}

		beers, total, err := repo.FilterAndPaginateBeers(ctx, name, page, limit)
		require.NoError(t, err)

		assert.True(t, total > 0)
		assert.Len(t, beers, int(limit))
		for _, beer := range beers {
			assert.Contains(t, beer.Name, name)
		}
	})

}
