package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

func LoadConfig(path string) IConfig {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("Error: .env file not found at path: %s", path)
	}

	// Read the .env file
	envMap, err := godotenv.Read(path)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	parseInt := func(key string) int {
		value, err := strconv.Atoi(envMap[key])
		if err != nil {
			log.Fatalf("load %s failed: %v", key, err)
		}
		return value
	}

	return &config{
		app: &app{
			host: envMap["APP_HOST"],
			port: envMap["APP_PORT"],
			name: envMap["APP_NAME"],
		},
		db: &db{
			mongodbUri:         envMap["MONGODB_URI"],
			mongodbName:        envMap["MONGODB_DB_NAME"],
			usersCollection:    envMap["USERS_COLLECTION"],
			productsCollection: envMap["PRODUCTS_COLLECTION"],
			signInsCollection:  envMap["USERS_SIGNIN_COLLECTION"],
			rolesCollection:    envMap["ROLES_COLLECTION"],
		},
		jwt: &jwt{
			adminKey:         envMap["JWT_ADMIN_KEY"],
			secretKey:        envMap["JWT_SECRET_KEY"],
			apiKey:           envMap["JWT_API_KEY"],
			accessExpiresAt:  parseInt("JWT_ACCESS_EXPIRES"),
			refreshExpiresAt: parseInt("JWT_REFRESH_EXPIRES"),
		},
	}
}

type IConfig interface {
	App() IAppConfig
	Db() IDbConfig
	Jwt() IJwtConfig
}

type config struct {
	app *app
	db  *db
	jwt *jwt
}

type app struct {
	host string
	port string
	name string
}

type db struct {
	mongodbUri         string
	mongodbName        string
	usersCollection    string
	productsCollection string
	signInsCollection  string
	rolesCollection    string
}

type jwt struct {
	adminKey         string
	secretKey        string
	apiKey           string
	accessExpiresAt  int //seconds
	refreshExpiresAt int //seconds
}

func (c *config) App() IAppConfig {
	return c.app
}

type IAppConfig interface {
	Url() string
	Port() string
	Host() string
	Name() string
}

func (a *app) Url() string  { return fmt.Sprintf("%s:%d", a.host, a.port) }
func (a *app) Port() string { return a.port }
func (a *app) Host() string { return a.host }
func (a *app) Name() string { return a.name }

func (c *config) Db() IDbConfig {
	return c.db
}

type IDbConfig interface {
	Url() string
	Name() string
	UsersCollection() string
	ProductsCollection() string
	SignInsCollection() string
	RolesCollection() string
}

func (d *db) Url() string {
	return d.mongodbUri
}

func (d *db) Name() string {
	return d.mongodbName
}

func (d *db) UsersCollection() string {
	return d.usersCollection
}

func (d *db) ProductsCollection() string {
	return d.productsCollection
}

func (d *db) SignInsCollection() string {
	return d.signInsCollection
}

func (d *db) RolesCollection() string {
	return d.rolesCollection
}

func (c *config) Jwt() IJwtConfig {
	return c.jwt
}

type IJwtConfig interface {
	AdminKey() string
	SecretKey() string
	ApiKey() string
	AccessExpiresAt() int
	RefreshExpiresAt() int
}

func (j *jwt) AdminKey() string      { return j.adminKey }
func (j *jwt) SecretKey() string     { return j.secretKey }
func (j *jwt) ApiKey() string        { return j.apiKey }
func (j *jwt) AccessExpiresAt() int  { return j.accessExpiresAt }
func (j *jwt) RefreshExpiresAt() int { return j.refreshExpiresAt }
