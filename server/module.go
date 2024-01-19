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
	usersGroup.POST("/refresh", userHandler.RefreshPassport)
	usersGroup.POST("/refresh-admin", userHandler.RefreshPassportAdmin)
	usersGroup.POST("/signup-admin", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2), userHandler.SignUpAdmin)
	//usersGroup.POST("/signup-admin", m.mid.JwtAuth(), m.mid.Authorize(2), userHandler.SignUpAdmin)
	usersGroup.POST("/signup-admin-no-middleware", userHandler.SignUpAdmin)
	usersGroup.POST("/create-role", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2), userHandler.CreateRole)

	usersGroup.GET("/:user_id", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: true, AllowAdmin: true}), m.mid.ParamCheck(), userHandler.GerUserProfile)
	usersGroup.GET("/get-all-user", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2), userHandler.GetAllUserProfile)
	usersGroup.PATCH("/update-role/:user_id", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2), userHandler.UpdateRole)
	usersGroup.GET("/admin/secret", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2), userHandler.GenerateAdminToken)
}

func (m *moduleFactory) ProductsModule() {
	productRepo := repository.NewproductsRepository(m.s.cfg, m.s.mongoDatabase)
	productUseCase := usecases.NewBeerService(productRepo)
	productHandler := handlers.NewBeerHandlers(productUseCase)
	beersGroup := m.r.Group("/beer")
	//m.mid.AuthorizeString("admin", "user")

	//userRole := 1      // binary "0001"
	//adminRole := 2     // binary "0010"
	//managerRole := 4   // binary "0100"
	//employeeRole := 8  // binary "1000"
	//customerRole := 16 // binary "00010000"
	//supervisorRole := 32 // binary "00100000"
	//directorRole := 64  // binary "01000000"
	//executiveRole := 128 // binary "10000000"

	beersGroup.POST("/create", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2, 4, 8), productHandler.CreateBeer)
	beersGroup.GET("/get/:id", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: true, AllowAdmin: true}), m.mid.Authorize(2, 4, 8), productHandler.GetBeer)
	beersGroup.PATCH("/update/:id", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2, 4, 8), productHandler.UpdateBeer)
	beersGroup.DELETE("/delete/:id", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: false, AllowAdmin: true}), m.mid.Authorize(2, 4, 8), productHandler.DeleteBeer)
	beersGroup.GET("/filter-and-paginate", m.mid.JwtAuth(middlewares.JwtAuthConfig{AllowCustomer: true, AllowAdmin: true}), productHandler.FilterAndPaginateBeers)
}
