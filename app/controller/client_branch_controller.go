package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type ClientBranchController interface {
	Upsert(c *gin.Context)
	Detail(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
}

type ClientBranchControllerImpl struct {
	svc service.ClientBranchService
}

func (a ClientBranchControllerImpl) Upsert(c *gin.Context) { a.svc.Upsert(c) }
func (a ClientBranchControllerImpl) Detail(c *gin.Context) { a.svc.Detail(c) }
func (a ClientBranchControllerImpl) List(c *gin.Context)   { a.svc.List(c) }
func (a ClientBranchControllerImpl) Delete(c *gin.Context) { a.svc.Delete(c) }

func ClientBranchControllerInit(s service.ClientBranchService) *ClientBranchControllerImpl {
	return &ClientBranchControllerImpl{svc: s}
}
