package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type MenuController interface {
	Upsert(c *gin.Context)
	Detail(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
}

type MenuControllerImpl struct {
	svc service.MenuService
}

func (a MenuControllerImpl) Upsert(c *gin.Context) { a.svc.Upsert(c) }
func (a MenuControllerImpl) Detail(c *gin.Context) { a.svc.Detail(c) }
func (a MenuControllerImpl) List(c *gin.Context)   { a.svc.List(c) }
func (a MenuControllerImpl) Delete(c *gin.Context) { a.svc.Delete(c) }

func MenuControllerInit(s service.MenuService) *MenuControllerImpl {
	return &MenuControllerImpl{svc: s}
}
