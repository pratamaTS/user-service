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
	Create(ctx *gin.Context)
	Detail(ctx *gin.Context)
	List(ctx *gin.Context)
	WarehouseApprove(ctx *gin.Context)
	DriverAccept(ctx *gin.Context)
	ReceiveDone(ctx *gin.Context)
}

type StockTransferServiceImpl struct {
	repo     repository.StockTransferRepository
	authRepo repository.AuthRepository
}

func NewStockTransferService(repo repository.StockTransferRepository, authRepo repository.AuthRepository) *StockTransferServiceImpl {
	return &StockTransferServiceImpl{repo: repo, authRepo: authRepo}
}

func (s *StockTransferServiceImpl) Create(ctx *gin.Context) {
	var req dto.StockTransferCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	if req.FromBranchUUID == req.ToBranchUUID {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, errors.New("from_branch_uuid cannot equal to_branch_uuid"))
		return
	}

	if req.DriverUUID == "" {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, errors.New("driver_uuid required"))
		return
	}

	tr := dao.StockTransfer{
		FromBranchUUID: req.FromBranchUUID,
		ToBranchUUID:   req.ToBranchUUID,
		DriverUUID:     req.DriverUUID,
		RequesterNote:  req.Notes,
		RequestedBy:    req.RequestedBy,
		Items:          []dao.StockTransferItem{},
	}

	// Items cuma product_uuid + qty, detail product diambil saat approve / atau bisa kamu enhance nanti
	for _, it := range req.Items {
		tr.Items = append(tr.Items, dao.StockTransferItem{
			ProductUUID: it.ProductUUID,
			Qty:         it.Qty,
		})
	}

	res, err := s.repo.CreateDraft(&tr)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to create stock transfer", http.StatusInternalServerError, err)
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
	res, err := s.repo.Detail(uuid)
	if err != nil {
		helpers.JsonErr[any](ctx, "not found", http.StatusNotFound, err)
		return
	}
	helpers.JsonOK(ctx, "success", res)
}

func (s *StockTransferServiceImpl) List(ctx *gin.Context) {
	var req dto.FilterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	res, err := s.repo.List(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to list stock transfer", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", res)
}

func (s *StockTransferServiceImpl) WarehouseApprove(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}

	var req dto.StockTransferWarehouseApproveReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}

	accessToken := ctx.GetString("access_token")
	if accessToken == "" {
		helpers.JsonErr[any](ctx, "missing access token", http.StatusBadRequest, errors.New("no bearer token"))
		return
	}

	profile, err := s.authRepo.ValidateToken(accessToken)
	if err != nil {
		helpers.JsonErr[any](ctx, "validate token failed", http.StatusUnauthorized, err)
		return
	}

	res, err := s.repo.WarehouseApprove(uuid, req.Notes, profile.UUID)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to approve (warehouse)", http.StatusBadRequest, err)
		return
	}
	helpers.JsonOK(ctx, "success", res)
}

func (s *StockTransferServiceImpl) DriverAccept(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}

	var req dto.StockTransferDriverAcceptReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}

	accessToken := ctx.GetString("access_token")
	if accessToken == "" {
		helpers.JsonErr[any](ctx, "missing access token", http.StatusBadRequest, errors.New("no bearer token"))
		return
	}

	profile, err := s.authRepo.ValidateToken(accessToken)
	if err != nil {
		helpers.JsonErr[any](ctx, "validate token failed", http.StatusUnauthorized, err)
		return
	}

	res, err := s.repo.DriverAccept(uuid, req.Notes, profile.UUID)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to accept (driver)", http.StatusBadRequest, err)
		return
	}
	helpers.JsonOK(ctx, "success", res)
}

func (s *StockTransferServiceImpl) ReceiveDone(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}

	var req dto.StockTransferReceiveReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}

	accessToken := ctx.GetString("access_token")
	if accessToken == "" {
		helpers.JsonErr[any](ctx, "missing access token", http.StatusBadRequest, errors.New("no bearer token"))
		return
	}

	profile, err := s.authRepo.ValidateToken(accessToken)
	if err != nil {
		helpers.JsonErr[any](ctx, "validate token failed", http.StatusUnauthorized, err)
		return
	}

	res, err := s.repo.ReceiveDone(uuid, req.Notes, profile.UUID)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to receive (cashier)", http.StatusBadRequest, err)
		return
	}
	helpers.JsonOK(ctx, "success", res)
}
