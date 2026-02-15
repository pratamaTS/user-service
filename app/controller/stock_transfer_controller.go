package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type StockTransferController interface {
	Request(c *gin.Context)
	Receive(c *gin.Context)
	Detail(c *gin.Context)
	List(c *gin.Context)
}

type StockTransferControllerImpl struct {
	svc service.StockTransferService
}

func (a StockTransferControllerImpl) Request(c *gin.Context) { a.svc.Request(c) }
func (a StockTransferControllerImpl) Receive(c *gin.Context) { a.svc.Receive(c) }
func (a StockTransferControllerImpl) Detail(c *gin.Context)  { a.svc.Detail(c) }
func (a StockTransferControllerImpl) List(c *gin.Context)    { a.svc.List(c) }

func StockTransferControllerInit(s service.StockTransferService) *StockTransferControllerImpl {
	return &StockTransferControllerImpl{svc: s}
}
