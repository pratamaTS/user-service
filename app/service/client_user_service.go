package service

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/domain/dao"
	"harjonan.id/user-service/app/domain/dto"
	"harjonan.id/user-service/app/helpers"
	"harjonan.id/user-service/app/repository"
)

type ClientUserService interface {
	Upsert(ctx *gin.Context)
	Detail(ctx *gin.Context)
	List(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type ClientUserServiceImpl struct {
	repo repository.ClientUserRepository
}

func NewClientUserService(repo repository.ClientUserRepository) *ClientUserServiceImpl {
	return &ClientUserServiceImpl{repo: repo}
}

func (s *ClientUserServiceImpl) Upsert(ctx *gin.Context) {
	var req dao.ClientUser
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	if req.ClientUUID == "" {
		helpers.JsonErr[any](ctx, "missing identifier", http.StatusBadRequest, errors.New("client uuid required"))
		return
	}
	if req.UserUUID == "" {
		helpers.JsonErr[any](ctx, "missing identifier", http.StatusBadRequest, errors.New("user uuid required"))
		return
	}
	res, err := s.repo.SaveClientUser(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to save client user", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", res)
}

func (s *ClientUserServiceImpl) Detail(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}
	data, err := s.repo.DetailClientUser(uuid)
	if err != nil {
		helpers.JsonErr[any](ctx, "not found", http.StatusNotFound, err)
		return
	}
	helpers.JsonOK(ctx, "success", data)
}

func (s *ClientUserServiceImpl) List(ctx *gin.Context) {
	var req dto.FilterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	data, err := s.repo.ListClientUser(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to list client user", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", data)
}

func (s *ClientUserServiceImpl) Delete(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}
	if err := s.repo.DeleteClientUser(uuid); err != nil {
		helpers.JsonErr[any](ctx, "failed to delete client user", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK[struct{}](ctx, "success", struct{}{})
}
