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
	Role     string             `bson:"role" json:"role"`
	RoleId   int                `bson:"role_id" json:"role_id"`
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
	Role   string             `bson:"role" json:"role"`
	RoleId int                `bson:"role_id" json:"role_id"`
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
	Role     string             `bson:"role" json:"role"`
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
	Role   string `json:"role"`
	RoleID string `json:"role_id"`
}

type RoleCreateRequest struct {
	RoleID string `json:"role_id"`
	Role   string `json:"role"`
}

type UserProfile struct {
	UserName                        string         `json:"userName"`
	Profile                         Profile        `json:"profile"`
	Email                           Email          `json:"email"`
	Phone                           Phone          `json:"phone"`
	Password                        string         `json:"password"`
	HashedPassword                  HashedPassword `json:"hashedPassword"`
	PasswordChangeRequired          bool           `json:"passwordChangeRequired"`
	RequestPasswordlessRegistration bool           `json:"requestPasswordlessRegistration"`
	OtpCode                         string         `json:"otpCode"`
	OrganizationId                  string         `json:"organizationId"`
}

type Profile struct {
	FirstName         string `json:"firstName"`
	LastName          string `json:"lastName"`
	NickName          string `json:"nickName"`
	DisplayName       string `json:"displayName"`
	PreferredLanguage string `json:"preferredLanguage"`
	Gender            string `json:"gender"`
}

type Email struct {
	Email           string `json:"email"`
	IsEmailVerified bool   `json:"isEmailVerified"`
}

type Phone struct {
	Phone           string `json:"phone"`
	IsPhoneVerified bool   `json:"isPhoneVerified"`
}

type HashedPassword struct {
	Value string `json:"value"`
}

type DeleteUserRequest struct {
	UserID string `json:"userId"`
	Reason string `json:"reason"`
}

type UserZitadel struct {
	Username       string
	HashedPassword string
	Email          string
}
