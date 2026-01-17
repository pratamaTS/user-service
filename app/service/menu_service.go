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

type MenuService interface {
	Upsert(ctx *gin.Context)
	Detail(ctx *gin.Context)
	List(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type MenuServiceImpl struct {
	repo       repository.MenuRepository
	parentRepo repository.ParentMenuRepository
}

func NewMenuService(repo repository.MenuRepository, parentRepo repository.ParentMenuRepository) *MenuServiceImpl {
	return &MenuServiceImpl{repo: repo, parentRepo: parentRepo}
}

func (s *MenuServiceImpl) Upsert(ctx *gin.Context) {
	var req dao.Menu
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	if req.Title == "" {
		helpers.JsonErr[any](ctx, "missing identifier", http.StatusBadRequest, errors.New("title required"))
		return
	}
	if req.ParentUUID != "" {
		if _, err := s.parentRepo.DetailParentMenu(req.ParentUUID); err != nil {
			helpers.JsonErr[any](ctx, "parent not found", http.StatusBadRequest, err)
			return
		}
	}

	res, err := s.repo.SaveMenu(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to save menu", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", res)
}

func (s *MenuServiceImpl) Detail(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}
	data, err := s.repo.DetailMenu(uuid)
	if err != nil {
		helpers.JsonErr[any](ctx, "not found", http.StatusNotFound, err)
		return
	}
	helpers.JsonOK(ctx, "success", data)
}

func (s *MenuServiceImpl) List(ctx *gin.Context) {
	var req dto.APIRequest[dto.FilterRequest]
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	data, err := s.repo.ListMenu(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to list menu", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", data)
}

func (s *MenuServiceImpl) Delete(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}
	if err := s.repo.DeleteMenu(uuid); err != nil {
		helpers.JsonErr[any](ctx, "failed to delete menu", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK[struct{}](ctx, "success", struct{}{})
}
