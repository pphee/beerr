package servers

import (
	"github.com/gin-gonic/gin"
	middlewaresHandler "pok92deng/middleware/middlewareHandlers"
	middlewaresRepositories "pok92deng/middleware/middlewareRepositories"
	middlewaresUsecases "pok92deng/middleware/middlewareUsecases"
	"pok92deng/users/usersHandlers"
	"pok92deng/users/usersRepositories"
	"pok92deng/users/usersUsecases"
)

type IModuleFactory interface {
	//MonitorModule()
	UsersModule()
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
	repository := middlewaresRepositories.MiddlewaresRepository(s.usersCollection)
	usecase := middlewaresUsecases.MiddlewaresRepository(repository)
	return middlewaresHandler.MiddlewaresRepository(s.cfg, usecase)
}

func (m *moduleFactory) UsersModule() {
	repository := usersRepositories.UsersRepository(m.s.usersCollection)
	usecase := usersUsecases.UsersUsecase(m.s.cfg, repository)
	handler := usersHandlers.UsersHandler(m.s.cfg, usecase)
	usersGroup := m.r.Group("/users")
	usersGroup.POST("/signup", handler.SignUpCustomer)
	usersGroup.POST("/signin", handler.SignIn)
	usersGroup.POST("/refresh", handler.RefreshPassport)
	usersGroup.POST("/signup-admin", m.mid.JwtAuthAdmin(), handler.SignUpAdmin)
	//usersGroup.POST("/signup-admin", m.mid.JwtAuth(), m.mid.Authorize(2), handler.SignUpAdmin)
	usersGroup.POST("/signup-adminnomiddelware", handler.SignUpAdmin)

	usersGroup.GET("/:user_id", m.mid.JwtAuth(), m.mid.ParamCheck(), handler.GerUserProfile)
	usersGroup.GET("/admin/secret", m.mid.JwtAuth(), m.mid.Authorize(2), handler.GenerateAdminToken)
}
