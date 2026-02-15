package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type ProductController interface {
	Upsert(c *gin.Context)
	Detail(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
	BulkUpload(c *gin.Context)
}

type ProductControllerImpl struct {
	svc service.ProductService
}

func (a ProductControllerImpl) Upsert(c *gin.Context)     { a.svc.Upsert(c) }
func (a ProductControllerImpl) Detail(c *gin.Context)     { a.svc.Detail(c) }
func (a ProductControllerImpl) List(c *gin.Context)       { a.svc.List(c) }
func (a ProductControllerImpl) Delete(c *gin.Context)     { a.svc.Delete(c) }
func (a ProductControllerImpl) BulkUpload(c *gin.Context) { a.svc.BulkUpload(c) }

func ProductControllerInit(s service.ProductService) *ProductControllerImpl {
	return &ProductControllerImpl{svc: s}
}
