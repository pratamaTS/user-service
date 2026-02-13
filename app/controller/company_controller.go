package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type CompanyController interface {
	Upsert(c *gin.Context)
	Detail(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
}

type CompanyControllerImpl struct {
	svc service.CompanyService
}

func (a CompanyControllerImpl) Upsert(c *gin.Context) { a.svc.Upsert(c) }
func (a CompanyControllerImpl) Detail(c *gin.Context) { a.svc.Detail(c) }
func (a CompanyControllerImpl) List(c *gin.Context)   { a.svc.List(c) }
func (a CompanyControllerImpl) Delete(c *gin.Context) { a.svc.Delete(c) }

func CompanyControllerInit(s service.CompanyService) *CompanyControllerImpl {
	return &CompanyControllerImpl{svc: s}
}
