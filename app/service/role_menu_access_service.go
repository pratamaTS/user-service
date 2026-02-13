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

type RoleMenuAccessService interface {
	Upsert(ctx *gin.Context)
	Detail(ctx *gin.Context)
	List(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type RoleMenuAccessServiceImpl struct {
	repo repository.RoleMenuAccessRepository
}

func NewRoleMenuAccessService(repo repository.RoleMenuAccessRepository) *RoleMenuAccessServiceImpl {
	return &RoleMenuAccessServiceImpl{repo: repo}
}

func (s *RoleMenuAccessServiceImpl) Upsert(ctx *gin.Context) {
	var req dao.RoleMenuAccess
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	if req.RoleUUID == "" {
		helpers.JsonErr[any](ctx, "missing identifier", http.StatusBadRequest, errors.New("role uuid required"))
		return
	}
	if len(req.AccesibleMenu) == 0 {
		helpers.JsonErr[any](ctx, "missing menu access", http.StatusBadRequest, errors.New("menu access required"))
		return
	}

	res, err := s.repo.SaveRoleMenuAccess(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to save role menu access", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", res)
}

func (s *RoleMenuAccessServiceImpl) Detail(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}
	data, err := s.repo.DetailRoleMenuAccess(uuid)
	if err != nil {
		helpers.JsonErr[any](ctx, "not found", http.StatusNotFound, err)
		return
	}
	helpers.JsonOK(ctx, "success", data)
}

func (s *RoleMenuAccessServiceImpl) List(ctx *gin.Context) {
	var req dto.FilterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	data, err := s.repo.ListRoleMenuAccess(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to list role menu access", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", data)
}

func (s *RoleMenuAccessServiceImpl) Delete(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}
	if err := s.repo.DeleteRoleMenuAccess(uuid); err != nil {
		helpers.JsonErr[any](ctx, "failed to delete role menu access", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK[struct{}](ctx, "success", struct{}{})
}
