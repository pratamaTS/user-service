package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type AttendanceController interface {
	Upsert(c *gin.Context)
	List(c *gin.Context)
}

type AttendanceControllerImpl struct {
	svc service.AttendanceService
}

func (a AttendanceControllerImpl) Upsert(c *gin.Context) { a.svc.Upsert(c) }
func (a AttendanceControllerImpl) List(c *gin.Context)   { a.svc.List(c) }

func AttendanceControllerInit(s service.AttendanceService) *AttendanceControllerImpl {
	return &AttendanceControllerImpl{svc: s}
}
