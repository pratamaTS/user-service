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

type StockTransferService interface {
	Request(ctx *gin.Context)
	Receive(ctx *gin.Context)
	Detail(ctx *gin.Context)
	List(ctx *gin.Context)
}

type StockTransferServiceImpl struct {
	repo repository.StockTransferRepository
}

func NewStockTransferService(repo repository.StockTransferRepository) *StockTransferServiceImpl {
	return &StockTransferServiceImpl{repo: repo}
}

func (s *StockTransferServiceImpl) Request(ctx *gin.Context) {
	var req dao.StockTransfer
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}

	if req.FromBranchUUID == "" || req.ToBranchUUID == "" {
		helpers.JsonErr[any](ctx, "missing branch", http.StatusBadRequest, errors.New("from_branch_uuid and to_branch_uuid required"))
		return
	}
	if req.FromBranchUUID == req.ToBranchUUID {
		helpers.JsonErr[any](ctx, "invalid branch", http.StatusBadRequest, errors.New("from and to branch must be different"))
		return
	}
	if len(req.Items) == 0 {
		helpers.JsonErr[any](ctx, "missing items", http.StatusBadRequest, errors.New("items required"))
		return
	}
	for _, it := range req.Items {
		if it.ProductUUID == "" {
			helpers.JsonErr[any](ctx, "invalid item", http.StatusBadRequest, errors.New("product_uuid required"))
			return
		}
		if it.Unit == "" {
			helpers.JsonErr[any](ctx, "invalid item", http.StatusBadRequest, errors.New("unit required"))
			return
		}
		if it.Qty <= 0 {
			helpers.JsonErr[any](ctx, "invalid item", http.StatusBadRequest, errors.New("qty must be > 0"))
			return
		}
	}

	res, err := s.repo.Create(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to request transfer", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", res)
}

func (s *StockTransferServiceImpl) Receive(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}

	// minimal: ambil dari header / token claim kalau kamu sudah punya helper
	receivedBy := ctx.GetString("user_uuid") // kalau middleware kamu set ini
	if receivedBy == "" {
		receivedBy = "SYSTEM"
	}

	res, err := s.repo.Receive(uuid, receivedBy)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to receive transfer", http.StatusBadRequest, err)
		return
	}
	helpers.JsonOK(ctx, "success", res)
}

func (s *StockTransferServiceImpl) Detail(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}
	data, err := s.repo.Detail(uuid)
	if err != nil {
		helpers.JsonErr[any](ctx, "not found", http.StatusNotFound, err)
		return
	}
	helpers.JsonOK(ctx, "success", data)
}

func (s *StockTransferServiceImpl) List(ctx *gin.Context) {
	var req dto.APIRequest[dto.FilterRequest]
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	data, err := s.repo.List(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to list transfer", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", data)
}
