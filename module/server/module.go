package servers

import (
	"github.com/gin-gonic/gin"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/http/middleware"
	"pok92deng/module/middleware"
	"pok92deng/module/middleware/middlewareHandlers"
	"pok92deng/module/middleware/middlewareRepositories"
	"pok92deng/module/middleware/middlewareUsecases"
	"pok92deng/module/product/productsHandlers"
	"pok92deng/module/product/productsRepositories"
	"pok92deng/module/product/productsUsecases"
	"pok92deng/module/users/usersHandlers"
	"pok92deng/module/users/usersRepositories"
	"pok92deng/module/users/usersUsecases"
)

type IModuleFactory interface {
	UsersModule()
	ProductsModule()
}

type moduleFactory struct {
	r   *gin.RouterGroup
	s   *server
	mid middlewaresHandler.IMiddlewaresHandler
	mw  *middleware.Interceptor[*oauth.IntrospectionContext]
}

func InitModule(r *gin.RouterGroup, s *server, mid middlewaresHandler.IMiddlewaresHandler, mw *middleware.Interceptor[*oauth.IntrospectionContext]) IModuleFactory {
	return &moduleFactory{
		r:   r,
		s:   s,
		mid: mid,
		mw:  mw,
	}
}

func InitMiddlewares(s *server) middlewaresHandler.IMiddlewaresHandler {
	middlewareRepository := middlewaresRepositories.MiddlewaresRepository(s.cfg, s.mongoDatabase)
	middlewareUseCase := middlewaresUsecases.MiddlewareUsecases(middlewareRepository)
	return middlewaresHandler.MiddlewaresHandler(s.cfg, middlewareUseCase, usersUsecases.UsersUsecase(s.cfg, usersRepositories.UsersRepository(s.cfg, s.mongoDatabase, s.connectZitadel)))
}

func (m *moduleFactory) UsersModule() {

	userRepo := usersRepositories.UsersRepository(m.s.cfg, m.s.mongoDatabase, m.s.connectZitadel)
	userUseCase := usersUsecases.UsersUsecase(m.s.cfg, userRepo)
	userHandler := usersHandlers.UsersHandler(m.s.cfg, userUseCase, m.mw)

	usersGroup := m.r.Group("/users")
	usersGroup.POST("/signup", userHandler.SignUpCustomer)
	usersGroup.POST("/sign-in", userHandler.SignIn)
	usersGroup.POST("/refresh", userHandler.RefreshPassport)
	usersGroup.POST("/refresh-admin", userHandler.RefreshPassportAdmin)
	usersGroup.POST("/signup-admin", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2), userHandler.SignUpAdmin)
	usersGroup.POST("/signup-admin-no-middleware", userHandler.SignUpAdmin)
	usersGroup.POST("/create-role", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2), userHandler.CreateRole)
	usersGroup.GET("/:user_id", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: true, AllowAdmin: true}), m.mid.ParamCheck(), userHandler.GerUserProfile)
	usersGroup.GET("/get-all-user", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2), userHandler.GetAllUserProfile)
	usersGroup.PATCH("/update-role/:user_id", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2), userHandler.UpdateRole)
	usersGroup.GET("/auth-ctx", userHandler.AuthCtx)

	//usersGroup.POST("/create-user-zitadel", userHandler.CreateUserZitadel)
	//zitadel
	zitadelGroup := usersGroup.Group("/zitadel")
	zitadelGroup.POST("/", userHandler.CreateUserProfile)
	zitadelGroup.DELETE("/:zitadel_id", userHandler.DeleteUserZitadel)
	zitadelGroup.GET("/:zitadel_id", userHandler.GetUserZitadel)
	zitadelGroup.POST("/import", userHandler.ImportUserToZitadel)
	zitadelGroup.POST("/:zitadel_id/lock", userHandler.LockUserInZitadel)
	zitadelGroup.POST("/:zitadel_id/unlock", userHandler.UnlockUserInZitadel)
	zitadelGroup.GET("/:zitadel_id/metadata/:key", userHandler.GetMetadata)

	//zitadel
	//, m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: true, AllowAdmin: true}), m.mid.Authorize(1, 2)
	//usersGroup.POST("/signup-admin", m.mid.JwtAuth(), m.mid.Authorize(2), userHandler.SignUpAdmin)
	//usersGroup.GET("/admin/secret", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2), userHandler.GenerateAdminToken)
}

func (m *moduleFactory) ProductsModule() {
	productRepo := repository.NewproductsRepository(m.s.cfg, m.s.mongoDatabase)
	productUseCase := usecases.NewBeerService(productRepo)
	productHandler := handlers.NewBeerHandlers(productUseCase)
	beersGroup := m.r.Group("/beer")

	//userRole := 1      // binary "0001"
	//adminRole := 2     // binary "0010"
	//managerRole := 4   // binary "0100"
	//employeeRole := 8  // binary "1000"
	//customerRole := 16 // binary "00010000"
	//supervisorRole := 32 // binary "00100000"
	//directorRole := 64  // binary "01000000"
	//executiveRole := 128 // binary "10000000"

	beersGroup.POST("/create", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2, 4, 8), productHandler.CreateBeer)
	beersGroup.GET("/get/:id", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: true, AllowAdmin: true}), m.mid.Authorize(1, 2, 4, 8), productHandler.GetBeer)
	beersGroup.PATCH("/update/:id", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2, 4, 8), productHandler.UpdateBeer)
	beersGroup.DELETE("/delete/:id", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2, 4, 8), productHandler.DeleteBeer)
	beersGroup.GET("/filter-and-paginate", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: true, AllowAdmin: true}), m.mid.Authorize(1, 2, 4, 8), productHandler.FilterAndPaginateBeers)
	beersGroup.GET("/filter-and-paginate-no-middleware", productHandler.FilterAndPaginateBeers)
}
