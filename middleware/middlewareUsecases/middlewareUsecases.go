package middlewaresUsecases

import (
	"context"
	middlewares "pok92deng/middleware"
	middlewaresRepositories "pok92deng/middleware/middlewareRepositories"
)

type IMiddlewaresUsecase interface {
	FindAccessToken(userId, accessToken string) bool
	FindRole(ctx context.Context, userRoleId string) ([]*middlewares.Roles, error)
}

type middlewaresUsecase struct {
	middlewaresRepository middlewaresRepositories.IMiddlewaresRepository
}

func MiddlewaresRepository(middlewaresRepository middlewaresRepositories.IMiddlewaresRepository) IMiddlewaresUsecase {
	return &middlewaresUsecase{
		middlewaresRepository: middlewaresRepository,
	}
}

func (u *middlewaresUsecase) FindAccessToken(userId, accessToken string) bool {
	return u.middlewaresRepository.FindAccessToken(userId, accessToken)
}

func (u *middlewaresUsecase) FindRole(ctx context.Context, userRoleId string) ([]*middlewares.Roles, error) {
	roles, err := u.middlewaresRepository.FindRole(ctx, userRoleId)
	if err != nil {
		return nil, err
	}
	return roles, nil
}
