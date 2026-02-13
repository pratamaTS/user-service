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

type CompanyService interface {
	Upsert(ctx *gin.Context)
	Detail(ctx *gin.Context)
	List(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type CompanyServiceImpl struct {
	repo repository.CompanyRepository
}

func NewCompanyService(repo repository.CompanyRepository) *CompanyServiceImpl {
	return &CompanyServiceImpl{repo: repo}
}

func (s *CompanyServiceImpl) Upsert(ctx *gin.Context) {
	var req dao.Company
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	if req.EmailCompany == "" {
		helpers.JsonErr[any](ctx, "missing identifier", http.StatusBadRequest, errors.New("email_company required"))
		return
	}
	res, err := s.repo.SaveCompany(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to save company", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", res)
}

func (s *CompanyServiceImpl) Detail(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}
	data, err := s.repo.DetailCompany(uuid)
	if err != nil {
		helpers.JsonErr[any](ctx, "not found", http.StatusNotFound, err)
		return
	}
	helpers.JsonOK(ctx, "success", data)
}

func (s *CompanyServiceImpl) List(ctx *gin.Context) {
	var req dto.APIRequest[dto.FilterRequest]
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	data, err := s.repo.ListCompany(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to list company", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", data)
}

func (s *CompanyServiceImpl) Delete(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}
	if err := s.repo.DeleteCompany(uuid); err != nil {
		helpers.JsonErr[any](ctx, "failed to delete company", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK[struct{}](ctx, "success", struct{}{})
}
