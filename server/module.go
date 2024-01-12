package servers

import (
	middlewares "pok92deng/middleware"
	middlewaresHandler "pok92deng/middleware/middlewareHandlers"
	middlewaresRepositories "pok92deng/middleware/middlewareRepositories"
	middlewaresUsecases "pok92deng/middleware/middlewareUsecases"
	handlers "pok92deng/product/productsHandlers"
	repository "pok92deng/product/productsRepositories"
	usecases "pok92deng/product/productsUsecases"
	"pok92deng/users/usersHandlers"
	"pok92deng/users/usersRepositories"
	"pok92deng/users/usersUsecases"

	"github.com/gin-gonic/gin"
)

type IModuleFactory interface {
	UsersModule()
	ProductsModule()
}

type moduleFactory struct {
	r   *gin.RouterGroup
	s   *server
	mid middlewaresHandler.IMiddlewaresHandler
}

func InitModule(r *gin.RouterGroup, s *server, mid middlewaresHandler.IMiddlewaresHandler) IModuleFactory {
	return &moduleFactory{
		r:   r,
		s:   s,
		mid: mid,
	}
}

func InitMiddlewares(s *server) middlewaresHandler.IMiddlewaresHandler {
	middlewareRepository := middlewaresRepositories.MiddlewaresRepository(s.cfg, s.mongoDatabase)
	middlewareUseCase := middlewaresUsecases.MiddlewaresRepository(middlewareRepository)
	return middlewaresHandler.MiddlewaresRepository(s.cfg, middlewareUseCase)
}

func (m *moduleFactory) UsersModule() {
	userRepo := usersRepositories.UsersRepository(m.s.cfg, m.s.mongoDatabase)
	userUseCase := usersUsecases.UsersUsecase(m.s.cfg, userRepo)
	userHandler := usersHandlers.UsersHandler(m.s.cfg, userUseCase)
	usersGroup := m.r.Group("/users")
	usersGroup.POST("/signup", userHandler.SignUpCustomer)
	usersGroup.POST("/sign-in", userHandler.SignIn)
	usersGroup.POST("/refresh", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: true, AllowAdmin: true}), userHandler.RefreshPassport)
	usersGroup.POST("/signup-admin", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), userHandler.SignUpAdmin)
	//usersGroup.POST("/signup-admin", m.mid.JwtAuth(), m.mid.Authorize(2), userHandler.SignUpAdmin)
	usersGroup.POST("/signup-admin-no-middleware", userHandler.SignUpAdmin)

	usersGroup.GET("/:user_id", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: true, AllowAdmin: true}), m.mid.ParamCheck(), userHandler.GerUserProfile)
	usersGroup.GET("/admin/secret", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2), userHandler.GenerateAdminToken)
}

func (m *moduleFactory) ProductsModule() {
	productRepo := repository.NewproductsRepository(m.s.cfg, m.s.mongoDatabase)
	productUseCase := usecases.NewBeerService(productRepo)
	productHandler := handlers.NewBeerHandlers(productUseCase)
	beersGroup := m.r.Group("/beer")

	beersGroup.POST("/create", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2), productHandler.CreateBeer)
	beersGroup.GET("/get/:id", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: true, AllowAdmin: true}), productHandler.GetBeer)
	beersGroup.PATCH("/update/:id", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2), productHandler.UpdateBeer)
	beersGroup.DELETE("/delete/:id", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2), productHandler.DeleteBeer)
	beersGroup.PATCH("/get-update/:id", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2), productHandler.GetUpdateBeer)
	beersGroup.GET("/filter-and-paginate", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: true, AllowAdmin: true}), productHandler.FilterAndPaginateBeers)
}
