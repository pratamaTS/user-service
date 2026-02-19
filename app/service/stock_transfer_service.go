package service

import (
	"errors"
	"fmt"
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
	repo      repository.StockTransferRepository
	authRepo  repository.AuthRepository
	notifRepo repository.NotificationRepository
}

func NewStockTransferService(repo repository.StockTransferRepository, authRepo repository.AuthRepository, notifRepo repository.NotificationRepository) *StockTransferServiceImpl {
	return &StockTransferServiceImpl{repo: repo, authRepo: authRepo, notifRepo: notifRepo}
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

	// ✅ Create notifications (gudang/from, driver personal, to branch)
	err = s.createStockTransferNotifs(profile.Client.UUID, res, "WAREHOUSE_APPROVED")
	if err != nil {
		fmt.Printf("failed to create notification: %v\n", err)
		// notifikasi gagal tidak kita anggap fatal, jadi tetap return success
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

	err = s.createStockTransferNotifs(profile.Client.UUID, res, "DRIVER_ACCEPTED")
	if err != nil {
		fmt.Printf("failed to create notification: %v\n", err)
		// notifikasi gagal tidak kita anggap fatal, jadi tetap return success
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

	err = s.createStockTransferNotifs(profile.Client.UUID, res, "RECEIVE_DONE")
	if err != nil {
		fmt.Printf("failed to create notification: %v\n", err)
		// notifikasi gagal tidak kita anggap fatal, jadi tetap return success
	}

	helpers.JsonOK(ctx, "success", res)
}

// =====================================================
// NOTIF BUILDER (3 target):
// - from_branch_uuid (gudang)
// - to_branch_uuid
// - driver_uuid as personal notif (user_uuid)
// Owner dapat semua notif via fetch client_uuid
// =====================================================
func (s *StockTransferServiceImpl) createStockTransferNotifs(clientUUID string, tr dao.StockTransfer, event string) error {
	if clientUUID == "" {
		return nil
	}

	ref := tr.UUID
	if ref == "" {
		ref = tr.UUID
	}

	var title, msg string
	icon := "info"

	switch event {
	case "WAREHOUSE_APPROVED":
		title = "Stock Transfer Approved"
		msg = fmt.Sprintf("Transfer #%s sudah di-approve. Silakan cek gudang tujuan.", shortRef(tr.UUID))
		icon = "success"
	case "DRIVER_ACCEPTED":
		title = "Driver Accepted Job"
		msg = fmt.Sprintf("Driver menerima job transfer #%s. Status: IN_PROGRESS.", shortRef(tr.UUID))
		icon = "info"
	case "RECEIVE_DONE":
		title = "Stock Transfer Received"
		msg = fmt.Sprintf("Transfer #%s sudah diterima oleh cabang tujuan. Status: DONE.", shortRef(tr.UUID))
		icon = "success"
	default:
		title = "Stock Transfer Update"
		msg = fmt.Sprintf("Update transfer #%s.", shortRef(tr.UUID))
		icon = "info"
	}

	// 1) From branch (gudang)
	if tr.FromBranchUUID != "" {
		_, _ = s.notifRepo.Insert(&dao.Notification{
			ClientUUID: clientUUID,
			BranchUUID: tr.FromBranchUUID,
			Title:      title,
			Message:    msg,
			Icon:       icon,
			Type:       "STOCK_TRANSFER",
			Ref:        tr.UUID,
		})
	}

	// 2) To branch (destination)
	if tr.ToBranchUUID != "" {
		_, _ = s.notifRepo.Insert(&dao.Notification{
			ClientUUID: clientUUID,
			BranchUUID: tr.ToBranchUUID,
			Title:      title,
			Message:    msg,
			Icon:       icon,
			Type:       "STOCK_TRANSFER",
			Ref:        tr.UUID,
		})
	}

	// 3) Driver personal (user_uuid)
	if tr.DriverUUID != "" {
		_, _ = s.notifRepo.Insert(&dao.Notification{
			ClientUUID: clientUUID,
			UserUUID:   tr.DriverUUID,
			Title:      title,
			Message:    msg,
			Icon:       icon,
			Type:       "STOCK_TRANSFER",
			Ref:        tr.UUID,
		})
	}

	return nil
}

func shortRef(u string) string {
	if u == "" {
		return "-"
	}
	if len(u) <= 8 {
		return u
	}
	return u[:8]
}
