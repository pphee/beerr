package auth

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"math"
	"pok92deng/config"
	"pok92deng/module/users"
	"time"
)

type TokenType string

const (
	Access            TokenType = "access"
	Refresh           TokenType = "refresh"
	Admin             TokenType = "admin"
	RefreshTokenAdmin TokenType = "refresh-token-admin"
)

type Auth struct {
	mapClaims *MapClaims
	cfg       config.IJwtConfig
}

type authAdmin struct {
	*Auth
}

type MapClaims struct {
	Claims *users.UserClaims `json:"claims"`
	jwt.RegisteredClaims
}

type IAuth interface {
	SignToken() string
}

type IAuthAdmin interface {
	SignToken() string
}

func jwtTimeDurationCal(t int) *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(time.Duration(int64(t) * int64(math.Pow10(9)))))
}

func jwtTimeRepeatAdapter(t int64) *jwt.NumericDate {
	return jwt.NewNumericDate(time.Unix(t, 0))
}

func (a *authAdmin) SignToken() string {
	secretAdminKey := []byte(a.cfg.AdminKey())
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, a.mapClaims)
	ss, _ := token.SignedString(secretAdminKey)
	return ss
}

func (a *Auth) SignToken() string {
	secretKey := []byte(a.cfg.SecretKey())
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, a.mapClaims)

	ss, err := token.SignedString(secretKey)
	if err != nil {
		fmt.Println("err", err)
		return ""
	}
	return ss
}

func ParseCustomerToken(cfg config.IJwtConfig, tokenString string) (*MapClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Method.Alg())
		}
		return []byte(cfg.SecretKey()), nil
	})

	if err != nil {
		fmt.Printf("Error parsing token: %v\n", err) // Additional logging
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, fmt.Errorf("token format is invalid")
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, fmt.Errorf("token had expired")
		default:
			return nil, fmt.Errorf("parse token failed: %v", err)
		}
	}

	if claims, ok := token.Claims.(*MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("claims type is invalid or token is not valid")
	}
}

func ParseAdminToken(cfg config.IJwtConfig, tokenString string) (*MapClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Method.Alg())
		}
		return []byte(cfg.AdminKey()), nil
	})

	if err != nil {
		fmt.Printf("Error parsing token: %v\n", err)
		return nil, err
	}

	if claims, ok := token.Claims.(*MapClaims); ok {
		if token.Valid {
			return claims, nil
		} else {
			fmt.Println("Token is not valid")
			return nil, fmt.Errorf("token is invalid")
		}
	} else {
		fmt.Println("Claims type is invalid")
		return nil, fmt.Errorf("claims type is invalid")
	}
}

func RepeatToken(cfg config.IJwtConfig, claims *users.UserClaims, exp int64) string {
	obj := &Auth{
		cfg: cfg,
		mapClaims: &MapClaims{
			Claims: claims,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "goEcommerce-api",
				Subject:   "refresh-token",
				Audience:  []string{"user"},
				ExpiresAt: jwtTimeRepeatAdapter(exp),
				NotBefore: jwt.NewNumericDate(time.Now()),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		},
	}
	return obj.SignToken()
}

func RepeatAdminToken(cfg config.IJwtConfig, claims *users.UserClaims, exp int64) string {
	obj := &authAdmin{
		Auth: &Auth{
			cfg: cfg,
			mapClaims: &MapClaims{
				Claims: claims,
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    "goEcommerce-api",
					Subject:   "refresh-admin-token",
					Audience:  []string{"admin"},
					ExpiresAt: jwtTimeRepeatAdapter(exp),
					NotBefore: jwt.NewNumericDate(time.Now()),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				},
			},
		},
	}
	return obj.SignToken()
}

func NewAuth(tokenType TokenType, cfg config.IJwtConfig, claims *users.UserClaims) (IAuth, error) {
	switch tokenType {
	case Access:
		return newAccessToken(cfg, claims), nil
	case Refresh:
		return newRefreshToken(cfg, claims), nil
	case Admin:
		return newAdminToken(cfg, claims), nil
	case RefreshTokenAdmin:
		return newRefreshTokenAdmin(cfg, claims), nil
	default:
		return nil, fmt.Errorf("unknown token type")
	}
}

func newAccessToken(cfg config.IJwtConfig, claims *users.UserClaims) IAuth {
	return &Auth{
		cfg: cfg,
		mapClaims: &MapClaims{
			Claims: claims,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "goEcommerce-api",
				Subject:   "access-token",
				Audience:  []string{claims.Role},
				ExpiresAt: jwtTimeDurationCal(cfg.AccessExpiresAt()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		},
	}
}

func newRefreshToken(cfg config.IJwtConfig, claims *users.UserClaims) IAuth {
	return &Auth{
		cfg: cfg,
		mapClaims: &MapClaims{
			Claims: claims,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "goEcommerce-api",
				Subject:   "refresh-token",
				Audience:  []string{claims.Role},
				ExpiresAt: jwtTimeDurationCal(cfg.RefreshExpiresAt()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		},
	}
}

func newRefreshTokenAdmin(cfg config.IJwtConfig, claims *users.UserClaims) IAuth {
	return &authAdmin{
		&Auth{
			cfg: cfg,
			mapClaims: &MapClaims{
				Claims: claims,
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    "goEcommerce-api",
					Subject:   "refresh-token",
					Audience:  []string{"admin"},
					ExpiresAt: jwtTimeDurationCal(cfg.RefreshExpiresAt()),
					NotBefore: jwt.NewNumericDate(time.Now()),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				},
			},
		},
	}
}

func newAdminToken(cfg config.IJwtConfig, claims *users.UserClaims) IAuth {
	return &authAdmin{
		Auth: &Auth{
			cfg: cfg,
			mapClaims: &MapClaims{
				Claims: claims,
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    "goEcommerce-api",
					Subject:   "admin-token",
					Audience:  []string{"admin"},
					ExpiresAt: jwtTimeDurationCal(300),
					NotBefore: jwt.NewNumericDate(time.Now()),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				},
			},
		},
	}
}
