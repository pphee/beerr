package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"strconv"
)

func LoadConfig(path string) IConfig {
	envMap, err := godotenv.Read(path)
	if err != nil {
		log.Fatalf("load dotenv failed: %v", err)
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
			mongodburi:         envMap["MONGODB_URI"],
			mongodbname:        envMap["MONGODB_DB_NAME"],
			userscollection:    envMap["USERS_COLLECTION"],
			productscollection: envMap["PRODUCTS_COLLECTION"],
			signinscollection:  envMap["SIGNINS_COLLECTION"],
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
	mongodburi         string
	mongodbname        string
	userscollection    string
	productscollection string
	signinscollection  string
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
	SigninsCollection() string
}

func (d *db) Url() string {
	return d.mongodburi
}

func (d *db) Name() string {
	return d.mongodbname
}

func (d *db) UsersCollection() string {
	return d.userscollection
}

func (d *db) ProductsCollection() string {
	return d.productscollection
}

func (d *db) SigninsCollection() string {
	return d.signinscollection
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
