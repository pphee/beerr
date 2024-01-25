package usersUsecases

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"pok92deng/config"
	"pok92deng/module/users"
	"pok92deng/module/users/usersRepositories"
	auth "pok92deng/pkg"
)

type IUsersUsecase interface {
	InsertCustomer(req *users.UserRegisterReq) (*users.UserPassport, error)
	GetPassport(req *users.UserCredential) (*users.UserPassport, error)
	RefreshPassport(req *users.UserRefreshCredential) (*users.UserPassport, error)
	GetUserProfile(userId string) (*users.User, error)
	InsertAdmin(req *users.UserRegisterReq) (*users.UserPassport, error)
	RefreshPassportAdmin(req *users.UserRefreshCredential) (*users.UserPassport, error)
	GetAllUserProfile() ([]*users.User, error)
	UpdateRole(userId string, roleId int, role string) error
	CreateRole(roleId, role string) error
}

type usersUsecase struct {
	cfg            config.IConfig
	userRepository usersRepositories.UserRepository
}

func UsersUsecase(cfg config.IConfig, userRepository usersRepositories.UserRepository) IUsersUsecase {
	return &usersUsecase{
		cfg:            cfg,
		userRepository: userRepository,
	}
}

func (u *usersUsecase) InsertCustomer(req *users.UserRegisterReq) (*users.UserPassport, error) {
	if err := req.BcryptHashing(); err != nil {
		return nil, err
	}
	result, err := u.userRepository.InsertUser(req, false)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (u *usersUsecase) InsertAdmin(req *users.UserRegisterReq) (*users.UserPassport, error) {
	if err := req.BcryptHashing(); err != nil {
		return nil, err
	}
	result, err := u.userRepository.InsertUser(req, true)
	if err != nil {
		return nil, err
	}

	return result, nil
}

const (
	AdminRoleId = 2
)

func (u *usersUsecase) GetPassport(req *users.UserCredential) (*users.UserPassport, error) {
	user, err := u.userRepository.FindOneUserByEmail(req.Email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	tokenType := auth.Access
	if user.RoleId == AdminRoleId {
		tokenType = auth.Admin
	}

	userForToken := &users.User{
		Id:       user.Id,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
		RoleId:   user.RoleId,
	}

	token, err := u.createToken(tokenType, userForToken)
	if err != nil {
		return nil, err
	}

	passport := &users.UserPassport{
		User: &users.User{
			Id:       user.Id,
			Email:    user.Email,
			Username: user.Username,
			Role:     user.Role,
			RoleId:   user.RoleId,
		},
		Token: token,
	}

	if err := u.userRepository.InsertOauth(passport); err != nil {
		return nil, fmt.Errorf("failed to insert OAuth data: %v", err)
	}

	return passport, nil
}

func (u *usersUsecase) createToken(tokenType auth.TokenType, user *users.User) (*users.UserToken, error) {
	accessToken, err := auth.NewAuth(tokenType, u.cfg.Jwt(), &users.UserClaims{
		Id:     user.Id,
		Role:   user.Role,
		RoleId: user.RoleId,
	})
	if err != nil {
		return nil, err
	}

	refreshTokenType := auth.Refresh
	if tokenType == auth.Admin {
		refreshTokenType = auth.RefreshTokenAdmin
	}

	refreshToken, err := auth.NewAuth(refreshTokenType, u.cfg.Jwt(), &users.UserClaims{
		Id:     user.Id,
		Role:   user.Role,
		RoleId: user.RoleId,
	})
	if err != nil {
		return nil, err
	}

	return &users.UserToken{
		AccessToken:  accessToken.SignToken(),
		RefreshToken: refreshToken.SignToken(),
	}, nil
}

func (u *usersUsecase) RefreshPassport(req *users.UserRefreshCredential) (*users.UserPassport, error) {
	claims, err := auth.ParseCustomerToken(u.cfg.Jwt(), req.RefreshToken)

	if err != nil {
		return nil, err
	}

	oauth, err := u.userRepository.FindOneOauth(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	profile, err := u.userRepository.GetProfile(oauth.UserId)
	if err != nil {
		return nil, err
	}

	newClaims := &users.UserClaims{
		Id:     profile.Id,
		Role:   profile.Role,
		RoleId: profile.RoleId,
	}

	accessToken, err := auth.NewAuth(
		auth.Access,
		u.cfg.Jwt(),
		newClaims,
	)
	if err != nil {
		return nil, err
	}

	refreshToken := auth.RepeatToken(
		u.cfg.Jwt(),
		newClaims,
		claims.ExpiresAt.Unix(),
	)

	passport := &users.UserPassport{
		User: profile,
		Token: &users.UserToken{
			Id:           oauth.Id,
			AccessToken:  accessToken.SignToken(),
			RefreshToken: refreshToken,
		},
	}

	if err := u.userRepository.UpdateOauth(passport.Token); err != nil {
		return nil, err
	}
	return passport, nil
}

func (u *usersUsecase) RefreshPassportAdmin(req *users.UserRefreshCredential) (*users.UserPassport, error) {
	claims, err := auth.ParseAdminToken(u.cfg.Jwt(), req.RefreshToken)

	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}

	oauth, err := u.userRepository.FindOneOauth(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	profile, err := u.userRepository.GetProfile(oauth.UserId)
	if err != nil {
		return nil, err
	}

	newClaims := &users.UserClaims{
		Id:     profile.Id,
		Role:   profile.Role,
		RoleId: profile.RoleId,
	}

	accessToken, err := auth.NewAuth(
		auth.Admin,
		u.cfg.Jwt(),
		newClaims,
	)
	if err != nil {
		return nil, err
	}

	refreshTokenAdmin := auth.RepeatAdminToken(
		u.cfg.Jwt(),
		newClaims,
		claims.ExpiresAt.Unix(),
	)

	passport := &users.UserPassport{
		User: profile,
		Token: &users.UserToken{
			Id:           oauth.Id,
			AccessToken:  accessToken.SignToken(),
			RefreshToken: refreshTokenAdmin,
		},
	}

	if err := u.userRepository.UpdateOauth(passport.Token); err != nil {
		return nil, err
	}
	return passport, nil
}

func (u *usersUsecase) GetUserProfile(userId string) (*users.User, error) {
	profile, err := u.userRepository.GetProfile(userId)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (u *usersUsecase) GetAllUserProfile() ([]*users.User, error) {
	profile, err := u.userRepository.GetAllUserProfile()
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (u *usersUsecase) UpdateRole(userId string, roleId int, role string) error {
	if err := u.userRepository.UpdateRole(userId, roleId, role); err != nil {
		return err
	}
	return nil
}

func (u *usersUsecase) CreateRole(roleId, role string) error {
	if err := u.userRepository.CreateRole(roleId, role); err != nil {
		return err
	}
	return nil
}
