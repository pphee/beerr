package databases

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockDbConfig struct {
	mock.Mock
}

func (m *MockDbConfig) Url() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDbConfig) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDbConfig) UsersCollection() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDbConfig) ProductsCollection() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDbConfig) SigninsCollection() string {
	args := m.Called()
	return args.String(0)
}

func TestConnectMongoDB(t *testing.T) {
	mockConfig := new(MockDbConfig)
	mockConfig.On("Url").Return("mongodb://localhost:27017")
	mockConfig.On("Name").Return("testdb")

	t.Run("SuccessfulConnection", func(t *testing.T) {
		db, err := ConnectMongoDB(mockConfig)
		assert.NoError(t, err)
		assert.NotNil(t, db)
	})
}
