package middlewares

type Role struct {
	Id    int    `db:"id"`
	Title string `db:"title"`
}

type JwtAuthConfig struct {
	AllowCustomer bool
	AllowAdmin    bool
}

type UsersRole string
