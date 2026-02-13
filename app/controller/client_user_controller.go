package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type ClientUserController interface {
	Upsert(c *gin.Context)
	Detail(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
}

type ClientUserControllerImpl struct {
	svc service.ClientUserService
}

func (a ClientUserControllerImpl) Upsert(c *gin.Context) { a.svc.Upsert(c) }
func (a ClientUserControllerImpl) Detail(c *gin.Context) { a.svc.Detail(c) }
func (a ClientUserControllerImpl) List(c *gin.Context)   { a.svc.List(c) }
func (a ClientUserControllerImpl) Delete(c *gin.Context) { a.svc.Delete(c) }

func ClientUserControllerInit(s service.ClientUserService) *ClientUserControllerImpl {
	return &ClientUserControllerImpl{svc: s}
}
