package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type UserController interface {
	Upsert(c *gin.Context)
	Detail(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
}

type UserControllerImpl struct {
	svc service.UserService
}

func (a UserControllerImpl) Upsert(c *gin.Context) { a.svc.Upsert(c) }
func (a UserControllerImpl) Detail(c *gin.Context) { a.svc.Detail(c) }
func (a UserControllerImpl) List(c *gin.Context)   { a.svc.List(c) }
func (a UserControllerImpl) Delete(c *gin.Context) { a.svc.Delete(c) }

func UserControllerInit(s service.UserService) *UserControllerImpl {
	return &UserControllerImpl{svc: s}
}
