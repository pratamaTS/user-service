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

type AdminService interface {
	Upsert(ctx *gin.Context)
	Detail(ctx *gin.Context)
	List(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type AdminServiceImpl struct {
	repo repository.AdminRepository
}

func NewAdminService(repo repository.AdminRepository) *AdminServiceImpl {
	return &AdminServiceImpl{repo: repo}
}

func (s *AdminServiceImpl) Upsert(ctx *gin.Context) {
	var req dao.Admin
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}

	if req.Email == "" {
		helpers.JsonErr[any](ctx, "missing identifier", http.StatusBadRequest, errors.New("email required"))
		return
	}

	result, err := s.repo.SaveAdmin(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to save admin", http.StatusInternalServerError, err)
		return
	}

	helpers.JsonOK(ctx, "success", result)
}

func (s *AdminServiceImpl) Detail(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}

	data, err := s.repo.DetailAdmin(uuid)
	if err != nil {
		helpers.JsonErr[any](ctx, "not found", http.StatusNotFound, err)
		return
	}

	helpers.JsonOK(ctx, "success", data)
}

func (s *AdminServiceImpl) List(ctx *gin.Context) {
	var req dto.APIRequest[dto.FilterRequest]
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}

	data, err := s.repo.ListAdmin(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to list admin", http.StatusInternalServerError, err)
		return
	}

	helpers.JsonOK(ctx, "success", data)
}

func (s *AdminServiceImpl) Delete(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}

	if err := s.repo.DeleteAdmin(uuid); err != nil {
		helpers.JsonErr[any](ctx, "failed to delete admin", http.StatusInternalServerError, err)
		return
	}

	helpers.JsonOK[struct{}](ctx, "success", struct{}{})
}
