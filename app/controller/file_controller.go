package controller

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/service"
)

type FileController interface {
	Upload(c *gin.Context)
	Get(c *gin.Context)
}

type FileControllerImpl struct {
	svc service.FileService
}

func (c FileControllerImpl) Upload(ctx *gin.Context) { c.svc.Upload(ctx) }
func (c FileControllerImpl) Get(ctx *gin.Context)    { c.svc.Get(ctx) }

func FileControllerInit(s service.FileService) *FileControllerImpl {
	return &FileControllerImpl{svc: s}
}
