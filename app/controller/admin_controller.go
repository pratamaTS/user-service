package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type AdminController interface {
	Upsert(c *gin.Context)
	Detail(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
}

type AdminControllerImpl struct {
	svc service.AdminService
}

func (a AdminControllerImpl) Upsert(c *gin.Context) { a.svc.Upsert(c) }
func (a AdminControllerImpl) Detail(c *gin.Context) { a.svc.Detail(c) }
func (a AdminControllerImpl) List(c *gin.Context)   { a.svc.List(c) }
func (a AdminControllerImpl) Delete(c *gin.Context) { a.svc.Delete(c) }

func AdminControllerInit(adminService service.AdminService) *AdminControllerImpl {
	return &AdminControllerImpl{svc: adminService}
}
