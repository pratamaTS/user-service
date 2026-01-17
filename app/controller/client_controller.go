package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type ClientController interface {
	Upsert(c *gin.Context)
	Detail(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
}

type ClientControllerImpl struct {
	svc service.ClientService
}

func (a ClientControllerImpl) Upsert(c *gin.Context) { a.svc.Upsert(c) }
func (a ClientControllerImpl) Detail(c *gin.Context) { a.svc.Detail(c) }
func (a ClientControllerImpl) List(c *gin.Context)   { a.svc.List(c) }
func (a ClientControllerImpl) Delete(c *gin.Context) { a.svc.Delete(c) }

func ClientControllerInit(s service.ClientService) *ClientControllerImpl {
	return &ClientControllerImpl{svc: s}
}
