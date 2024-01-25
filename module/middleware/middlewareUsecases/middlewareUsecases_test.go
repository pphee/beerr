package middlewaresUsecases

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"pok92deng/module/middleware"
	"testing"
)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) FindAccessToken(userId, accessToken string) bool {
	args := m.Called(userId, accessToken)
	return args.Bool(0)
}

func (m *mockRepository) FindRole(ctx context.Context, userRoleId string) ([]*middlewares.Roles, error) {
	args := m.Called(ctx, userRoleId)
	return args.Get(0).([]*middlewares.Roles), args.Error(1)
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
		ctx := context.Background()
		testRole := &middlewares.Roles{RoleID: "1", Role: "TestRole"}
		mockRepo.On("FindRole", ctx, "1").Return([]*middlewares.Roles{testRole}, nil)

		result, err := useCase.FindRole(ctx, "1")
		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "1", result[0].RoleID)
		assert.Equal(t, "TestRole", result[0].Role)
	})

	t.Run("FindRole_Error", func(t *testing.T) {
		expectedError := errors.New("find role error")
		ctx := context.Background()
		mockRepo.On("FindRole", ctx, "userRoleId").Return([]*middlewares.Roles(nil), expectedError).Once()
		result, err := useCase.FindRole(ctx, "userRoleId")
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})

}
