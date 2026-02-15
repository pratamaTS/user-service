package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type POSTransactionController interface {
	Checkout(c *gin.Context)
	Detail(c *gin.Context)
	List(c *gin.Context)
	Void(c *gin.Context)
	ScanByBarcode(c *gin.Context)
}

type POSTransactionControllerImpl struct {
	svc service.POSTransactionService
}

func (a POSTransactionControllerImpl) Checkout(c *gin.Context)      { a.svc.Checkout(c) }
func (a POSTransactionControllerImpl) Detail(c *gin.Context)        { a.svc.Detail(c) }
func (a POSTransactionControllerImpl) List(c *gin.Context)          { a.svc.List(c) }
func (a POSTransactionControllerImpl) Void(c *gin.Context)          { a.svc.Void(c) }
func (a POSTransactionControllerImpl) ScanByBarcode(c *gin.Context) { a.svc.ScanByBarcode(c) }

func POSTransactionControllerInit(s service.POSTransactionService) *POSTransactionControllerImpl {
	return &POSTransactionControllerImpl{svc: s}
}
