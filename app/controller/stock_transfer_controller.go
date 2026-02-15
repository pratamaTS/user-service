package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type StockTransferController interface {
	Create(c *gin.Context)
	Detail(c *gin.Context)
	List(c *gin.Context)
	WarehouseApprove(c *gin.Context)
	DriverAccept(c *gin.Context)
	ReceiveDone(c *gin.Context)
}

type StockTransferControllerImpl struct {
	svc service.StockTransferService
}

func (a StockTransferControllerImpl) Create(c *gin.Context)           { a.svc.Create(c) }
func (a StockTransferControllerImpl) Detail(c *gin.Context)           { a.svc.Detail(c) }
func (a StockTransferControllerImpl) List(c *gin.Context)             { a.svc.List(c) }
func (a StockTransferControllerImpl) WarehouseApprove(c *gin.Context) { a.svc.WarehouseApprove(c) }
func (a StockTransferControllerImpl) DriverAccept(c *gin.Context)     { a.svc.DriverAccept(c) }
func (a StockTransferControllerImpl) ReceiveDone(c *gin.Context)      { a.svc.ReceiveDone(c) }

func StockTransferControllerInit(s service.StockTransferService) *StockTransferControllerImpl {
	return &StockTransferControllerImpl{svc: s}
}
