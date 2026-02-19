package controller

import (
	"github.com/gin-gonic/gin"

	"harjonan.id/user-service/app/service"
)

type NotificationController interface {
	Fetch(ctx *gin.Context)
	MarkRead(ctx *gin.Context)
	MarkReadAll(ctx *gin.Context)
	Clear(ctx *gin.Context)
	ClearAll(ctx *gin.Context)
}

type NotificationControllerImpl struct {
	svc service.NotificationService
}

func NotificationControllerInit(svc service.NotificationService) *NotificationControllerImpl {
	return &NotificationControllerImpl{svc: svc}
}

func (c *NotificationControllerImpl) Fetch(ctx *gin.Context)       { c.svc.Fetch(ctx) }
func (c *NotificationControllerImpl) MarkRead(ctx *gin.Context)    { c.svc.MarkRead(ctx) }
func (c *NotificationControllerImpl) MarkReadAll(ctx *gin.Context) { c.svc.MarkReadAll(ctx) }
func (c *NotificationControllerImpl) Clear(ctx *gin.Context)       { c.svc.Clear(ctx) }
func (c *NotificationControllerImpl) ClearAll(ctx *gin.Context)    { c.svc.ClearAll(ctx) }
