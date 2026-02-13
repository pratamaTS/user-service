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

type UserService interface {
	Upsert(ctx *gin.Context)
	Detail(ctx *gin.Context)
	List(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type UserServiceImpl struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserServiceImpl {
	return &UserServiceImpl{repo: repo}
}

func (s *UserServiceImpl) Upsert(ctx *gin.Context) {
	var req dao.User
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	if req.Name == "" {
		helpers.JsonErr[any](ctx, "missing identifier", http.StatusBadRequest, errors.New("name required"))
		return
	}
	res, err := s.repo.SaveUser(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to save user", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", res)
}

func (s *UserServiceImpl) Detail(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}
	data, err := s.repo.DetailUser(uuid)
	if err != nil {
		helpers.JsonErr[any](ctx, "not found", http.StatusNotFound, err)
		return
	}
	helpers.JsonOK(ctx, "success", data)
}

func (s *UserServiceImpl) List(ctx *gin.Context) {
	var req dto.FilterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	data, err := s.repo.ListUser(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to list user", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", data)
}

func (s *UserServiceImpl) Delete(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}
	if err := s.repo.DeleteUser(uuid); err != nil {
		helpers.JsonErr[any](ctx, "failed to delete user", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK[struct{}](ctx, "success", struct{}{})
}
