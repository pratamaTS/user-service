package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type DashboardController interface {
	OwnerSummary(c *gin.Context)
}

type DashboardControllerImpl struct {
	svc service.DashboardService
}

func (a DashboardControllerImpl) OwnerSummary(c *gin.Context) { a.svc.OwnerSummary(c) }

func DashboardControllerInit(s service.DashboardService) *DashboardControllerImpl {
	return &DashboardControllerImpl{svc: s}
}
