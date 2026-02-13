package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type ParentMenuController interface {
	Upsert(c *gin.Context)
	Detail(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
}

type ParentMenuControllerImpl struct {
	svc service.ParentMenuService
}

func (a ParentMenuControllerImpl) Upsert(c *gin.Context) { a.svc.Upsert(c) }
func (a ParentMenuControllerImpl) Detail(c *gin.Context) { a.svc.Detail(c) }
func (a ParentMenuControllerImpl) List(c *gin.Context)   { a.svc.List(c) }
func (a ParentMenuControllerImpl) Delete(c *gin.Context) { a.svc.Delete(c) }

func ParentMenuControllerInit(s service.ParentMenuService) *ParentMenuControllerImpl {
	return &ParentMenuControllerImpl{svc: s}
}
