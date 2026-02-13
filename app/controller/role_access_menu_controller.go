package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type RoleAccessMenuController interface {
	Upsert(c *gin.Context)
	Detail(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
}

type RoleAccessMenuControllerImpl struct {
	svc service.RoleMenuAccessService
}

func (a RoleAccessMenuControllerImpl) Upsert(c *gin.Context) { a.svc.Upsert(c) }
func (a RoleAccessMenuControllerImpl) Detail(c *gin.Context) { a.svc.Detail(c) }
func (a RoleAccessMenuControllerImpl) List(c *gin.Context)   { a.svc.List(c) }
func (a RoleAccessMenuControllerImpl) Delete(c *gin.Context) { a.svc.Delete(c) }

func RoleAccessMenuControllerInit(s service.RoleMenuAccessService) *RoleAccessMenuControllerImpl {
	return &RoleAccessMenuControllerImpl{svc: s}
}
