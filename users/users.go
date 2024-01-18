package users

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"regexp"
)

type User struct {
	Id       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email    string             `bson:"email" json:"email"`
	Username string             `bson:"username" json:"username"`
	RoleId   int                `bson:"role_id" json:"role_id"` // Ensure this is an integer.
}

type UserPassport struct {
	User  *User      `json:"user"`
	Token *UserToken `json:"token,omitempty"`
}

type UserToken struct {
	Id           string `bson:"_id,omitempty" json:"id"`
	AccessToken  string `bson:"access_token" json:"access_token"`
	RefreshToken string `bson:"refresh_token" json:"refresh_token"`
}

type UserRegisterReq struct {
	Email    string `bson:"email" json:"email" form:"email"`
	Password string `bson:"password" json:"password" form:"password"`
	Username string `bson:"username" json:"username" form:"username"`
}

func (obj *UserRegisterReq) BcryptHashing() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(obj.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashed password failed: %v", err)
	}
	obj.Password = string(hashedPassword)
	return nil
}

func (obj *UserRegisterReq) IsEmail() bool {
	match, err := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, obj.Email)
	if err != nil {
		return false
	}
	return match
}

func (obj *UserRegisterReq) IsPassword() bool {
	pattern := `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[^a-zA-Z\d]).{8,}$`
	match, err := regexp.MatchString(pattern, obj.Password)
	if err != nil {
		return false
	}
	return match
}

type UserClaims struct {
	Id     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoleId int                `bson:"role" json:"role"`
}

type UserCredential struct {
	Email    string `bson:"email" json:"email" form:"email"`
	Password string `bson:"password" json:"password" form:"password"`
}

// UserCredentialCheck represents detailed user credentials, including the user ID, username, and role.
type UserCredentialCheck struct {
	Id       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email    string             `bson:"email" json:"email"`
	Password string             `bson:"password" json:"password"`
	Username string             `bson:"username" json:"username"`
	RoleId   int                `bson:"role_id" json:"role_id"`
}

type UserRefreshCredential struct {
	RefreshToken string `json:"refresh_token" form:"refresh_token"`
}

type Oauth struct {
	Id     string `db:"id" json:"id"`
	UserId string `db:"user_id" json:"user_id"`
}

type MongoPassport struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserID       primitive.ObjectID `bson:"user_id"`
	AccessToken  string             `bson:"access_token"`
	RefreshToken string             `bson:"refresh_token"`
}

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return e.Message
}

func NewHTTPError(statusCode int, message string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
	}
}

type RoleUpdateRequest struct {
	RoleID string `json:"role_id"`
}
