package middlewaresUsecases

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	middlewares "pok92deng/middleware"
	"testing"
)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) FindAccessToken(userId, accessToken string) bool {
	args := m.Called(userId, accessToken)
	return args.Bool(0)
}

func (m *mockRepository) FindRole() ([]*middlewares.Role, error) {
	args := m.Called()
	return args.Get(0).([]*middlewares.Role), args.Error(1)
}

func TestMiddlewaresRepository(t *testing.T) {
	mockRepo := new(mockRepository)
	useCase := MiddlewaresRepository(mockRepo)

	t.Run("FindAccessToken", func(t *testing.T) {
		mockRepo.On("FindAccessToken", "userId", "accessToken").Return(true)
		result := useCase.FindAccessToken("userId", "accessToken")
		assert.Equal(t, true, result)
	})

	t.Run("FindRole_Success", func(t *testing.T) {
		mockRepo.On("FindRole").Return([]*middlewares.Role{}, nil).Once()
		result, err := useCase.FindRole()
		assert.Equal(t, []*middlewares.Role{}, result)
		assert.Nil(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("FindRole_Error", func(t *testing.T) {
		expectedError := errors.New("find role error")
		mockRepo.On("FindRole").Return(([]*middlewares.Role)(nil), expectedError).Once()
		result, err := useCase.FindRole()
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})
}
