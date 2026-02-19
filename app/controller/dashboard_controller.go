package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type DashboardController interface {
	OwnerSummary(c *gin.Context)
	KasirSummary(c *gin.Context)
	GudangSummary(c *gin.Context)
	DriverSummary(c *gin.Context)
}

type DashboardControllerImpl struct {
	svc service.DashboardService
}

func (a DashboardControllerImpl) OwnerSummary(c *gin.Context)  { a.svc.OwnerSummary(c) }
func (a DashboardControllerImpl) KasirSummary(c *gin.Context)  { a.svc.KasirSummary(c) }
func (a DashboardControllerImpl) GudangSummary(c *gin.Context) { a.svc.GudangSummary(c) }
func (a DashboardControllerImpl) DriverSummary(c *gin.Context) { a.svc.DriverSummary(c) }

func DashboardControllerInit(s service.DashboardService) *DashboardControllerImpl {
	return &DashboardControllerImpl{svc: s}
}
