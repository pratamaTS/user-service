package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type SubscriptionController interface {
	Upsert(c *gin.Context)
	Detail(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
	ActivateFromMaster(c *gin.Context)
}

type SubscriptionControllerImpl struct {
	svc service.SubscriptionService
}

func (a SubscriptionControllerImpl) Upsert(c *gin.Context)             { a.svc.Upsert(c) }
func (a SubscriptionControllerImpl) Detail(c *gin.Context)             { a.svc.Detail(c) }
func (a SubscriptionControllerImpl) List(c *gin.Context)               { a.svc.List(c) }
func (a SubscriptionControllerImpl) Delete(c *gin.Context)             { a.svc.Delete(c) }
func (a SubscriptionControllerImpl) ActivateFromMaster(c *gin.Context) { a.svc.ActivateFromMaster(c) }

func SubscriptionControllerInit(s service.SubscriptionService) *SubscriptionControllerImpl {
	return &SubscriptionControllerImpl{svc: s}
}
