package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type RoleController interface {
	Upsert(c *gin.Context)
	Detail(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
}

type RoleControllerImpl struct {
	svc service.RoleService
}

func (a RoleControllerImpl) Upsert(c *gin.Context) { a.svc.Upsert(c) }
func (a RoleControllerImpl) Detail(c *gin.Context) { a.svc.Detail(c) }
func (a RoleControllerImpl) List(c *gin.Context)   { a.svc.List(c) }
func (a RoleControllerImpl) Delete(c *gin.Context) { a.svc.Delete(c) }

func RoleControllerInit(s service.RoleService) *RoleControllerImpl {
	return &RoleControllerImpl{svc: s}
}
