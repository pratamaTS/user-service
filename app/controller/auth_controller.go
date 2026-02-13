package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type AuthController interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	Refresh(c *gin.Context)
	Logout(c *gin.Context)
	Me(c *gin.Context)
}

type AuthControllerImpl struct {
	svc service.AuthService
}

func (a AuthControllerImpl) Register(c *gin.Context) { a.svc.Register(c) }
func (a AuthControllerImpl) Login(c *gin.Context)    { a.svc.Login(c) }
func (a AuthControllerImpl) Refresh(c *gin.Context)  { a.svc.Refresh(c) }
func (a AuthControllerImpl) Logout(c *gin.Context)   { a.svc.Logout(c) }
func (a AuthControllerImpl) Me(c *gin.Context)       { a.svc.Me(c) }

func AuthControllerInit(authService service.AuthService) *AuthControllerImpl {
	return &AuthControllerImpl{svc: authService}
}
