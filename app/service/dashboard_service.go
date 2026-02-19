package service

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"harjonan.id/user-service/app/domain/dao"
	"harjonan.id/user-service/app/domain/dto"
	"harjonan.id/user-service/app/helpers"
	"harjonan.id/user-service/app/repository"
)

type DashboardService interface {
	OwnerSummary(ctx *gin.Context)
	KasirSummary(ctx *gin.Context)
	GudangSummary(ctx *gin.Context)
	DriverSummary(ctx *gin.Context)
}

type DashboardServiceImpl struct {
	repo      repository.DashboardRepository
	authRepo  repository.AuthRepository
	notifRepo repository.NotificationRepository
}

func NewDashboardService(repo repository.DashboardRepository, authRepo repository.AuthRepository, notifRepo repository.NotificationRepository) *DashboardServiceImpl {
	return &DashboardServiceImpl{repo: repo, authRepo: authRepo, notifRepo: notifRepo}
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
func endOfDay(t time.Time) time.Time { return startOfDay(t).Add(24 * time.Hour) }
func startOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}
func endOfMonth(t time.Time) time.Time { return startOfMonth(t).AddDate(0, 1, 0) }

func (s *DashboardServiceImpl) pushLowStockNotification(
	clientUUID string,
	branchUUID string,
	product dao.Product,
) {

	title := "Low Stock Alert"
	message := product.Name + " stock tersisa " + strconv.FormatInt(product.Stock, 10)

	notif := &dao.Notification{
		ClientUUID: clientUUID,
		BranchUUID: branchUUID,
		Title:      title,
		Message:    message,
		Type:       "STOCK",
		Icon:       "warning",
		Ref:        product.UUID,
	}

	_, _ = s.notifRepo.Insert(notif)
}

// =========================
// OWNER (existing)
// =========================
func (s *DashboardServiceImpl) OwnerSummary(ctx *gin.Context) {
	var req dto.OwnerDashboardRequest
	_ = ctx.ShouldBindJSON(&req)

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

	clientUUID := profile.Client.UUID
	if clientUUID == "" {
		helpers.JsonErr[any](ctx, "client uuid not found", http.StatusUnauthorized, errors.New("missing client_uuid"))
		return
	}

	threshold := req.LowStockThreshold
	if threshold <= 0 {
		threshold = 10
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	dayStart := startOfDay(now)
	dayEnd := endOfDay(now)
	monthStart := startOfMonth(now)
	monthEnd := endOfMonth(now)

	branchUUIDs, err := s.repo.GetCompanyBranchUUIDs(clientUUID)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to get branches", http.StatusInternalServerError, err)
		return
	}

	totalProduct, err := s.repo.CountActiveProducts(clientUUID)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count products", http.StatusInternalServerError, err)
		return
	}

	totalStockReq, err := s.repo.CountStockRequestMonth(branchUUIDs, monthStart, monthEnd)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count stock request", http.StatusInternalServerError, err)
		return
	}

	stockReqProcess, err := s.repo.CountStockRequestProcess(branchUUIDs)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count stock request process", http.StatusInternalServerError, err)
		return
	}

	jobWaitingAccept, err := s.repo.CountJobWaitingAccept(branchUUIDs)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count job waiting accept", http.StatusInternalServerError, err)
		return
	}

	txToday, err := s.repo.CountTransactionToday(branchUUIDs, dayStart, dayEnd)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count transaction today", http.StatusInternalServerError, err)
		return
	}

	revenueMonth, err := s.repo.SumRevenueMonth(branchUUIDs, monthStart, monthEnd)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to sum revenue month", http.StatusInternalServerError, err)
		return
	}

	lowStockSKU, err := s.repo.CountLowStockSKU(branchUUIDs, threshold)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count low stock sku", http.StatusInternalServerError, err)
		return
	}

	driverAvailable, err := s.repo.CountDriverAvailable(clientUUID)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count driver available", http.StatusInternalServerError, err)
		return
	}

	if lowStockSKU > 0 {
		products, err := s.repo.FindLowStockProducts(branchUUIDs, int(threshold))
		if err == nil {
			for _, p := range products {
				s.pushLowStockNotification(clientUUID, p.BranchUUID, p)
			}
		}
	}

	resp := dto.OwnerDashboardSummary{
		TotalProduct:        totalProduct,
		TotalStockRequest:   totalStockReq,
		StockRequestProcess: stockReqProcess,
		DriverAvailable:     driverAvailable,
		JobWaitingAccept:    jobWaitingAccept,
		TransactionToday:    txToday,
		RevenueMonth:        revenueMonth,
		LowStockSKU:         lowStockSKU,
	}

	helpers.JsonOK(ctx, "success", resp)
}

// =========================
// KASIR SUMMARY
// =========================
func (s *DashboardServiceImpl) KasirSummary(ctx *gin.Context) {
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

	clientUUID := profile.Client.UUID
	branchUUID := profile.Branch.UUID // ✅ penting: kasir scope per-branch
	if clientUUID == "" || branchUUID == "" {
		helpers.JsonErr[any](ctx, "branch/client not found", http.StatusUnauthorized, errors.New("missing client_uuid/branch_uuid"))
		return
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	dayStart := startOfDay(now)
	dayEnd := endOfDay(now)
	monthStart := startOfMonth(now)
	monthEnd := endOfMonth(now)

	txToday, err := s.repo.CountTransactionToday([]string{branchUUID}, dayStart, dayEnd)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count transaction today", http.StatusInternalServerError, err)
		return
	}

	txMonth, err := s.repo.CountTransactionMonth(branchUUID, monthStart, monthEnd)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count transaction month", http.StatusInternalServerError, err)
		return
	}

	productInBranch, err := s.repo.CountActiveProductsByBranch(branchUUID)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count product in branch", http.StatusInternalServerError, err)
		return
	}

	stockReqProcess, err := s.repo.CountStockRequestProcess([]string{branchUUID})
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count stock request process", http.StatusInternalServerError, err)
		return
	}

	threshold := int64(10)

	lowStockSKU, err := s.repo.CountLowStockSKU([]string{branchUUID}, threshold)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count low stock sku", http.StatusInternalServerError, err)
		return
	}

	if lowStockSKU > 0 {
		products, err := s.repo.FindLowStockProducts([]string{branchUUID}, int(threshold))
		if err == nil {
			for _, p := range products {
				s.pushLowStockNotification(clientUUID, p.BranchUUID, p)
			}
		}
	}

	resp := dto.KasirDashboardSummary{
		TransactionToday:      txToday,
		TransactionMonth:      txMonth,
		ProductInBranch:       productInBranch,
		StockRequestInProcess: stockReqProcess,
	}

	helpers.JsonOK(ctx, "success", resp)
}

// =========================
// GUDANG SUMMARY
// =========================
func (s *DashboardServiceImpl) GudangSummary(ctx *gin.Context) {
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

	clientUUID := profile.Client.UUID
	if clientUUID == "" {
		helpers.JsonErr[any](ctx, "client uuid not found", http.StatusUnauthorized, errors.New("missing client_uuid"))
		return
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	monthStart := startOfMonth(now)
	monthEnd := endOfMonth(now)

	// ✅ gudang branch uuid by name "Gudang"
	gudangBranchUUID, err := s.repo.FindGudangBranchUUID(clientUUID)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to get gudang branch", http.StatusInternalServerError, err)
		return
	}

	totalProduct, err := s.repo.CountActiveProducts(clientUUID) // repo kamu sudah force gudang
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count products", http.StatusInternalServerError, err)
		return
	}

	totalStockReq, err := s.repo.CountStockRequestMonth([]string{gudangBranchUUID}, monthStart, monthEnd)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count stock request month", http.StatusInternalServerError, err)
		return
	}

	stockReqInDriver, err := s.repo.CountStockRequestInDriver([]string{gudangBranchUUID})
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count stock request in driver", http.StatusInternalServerError, err)
		return
	}

	driverAvailable, err := s.repo.CountDriverAvailable(clientUUID)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count driver available", http.StatusInternalServerError, err)
		return
	}

	productSent, err := s.repo.SumProductSentToBranchMonth(gudangBranchUUID, monthStart, monthEnd)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to sum product sent", http.StatusInternalServerError, err)
		return
	}

	threshold := int64(10)

	lowStockSKU, err := s.repo.CountLowStockSKU([]string{gudangBranchUUID}, threshold)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count low stock sku", http.StatusInternalServerError, err)
		return
	}

	if lowStockSKU > 0 {
		products, err := s.repo.FindLowStockProducts([]string{gudangBranchUUID}, int(threshold))
		if err == nil {
			for _, p := range products {
				s.pushLowStockNotification(clientUUID, p.BranchUUID, p)
			}
		}
	}

	resp := dto.GudangDashboardSummary{
		TotalProduct:           totalProduct,
		TotalStockRequestMonth: totalStockReq,
		StockRequestInDriver:   stockReqInDriver,
		DriverAvailable:        driverAvailable,
		ProductSentToBranch:    productSent,
	}

	helpers.JsonOK(ctx, "success", resp)
}

// =========================
// DRIVER SUMMARY
// =========================
func (s *DashboardServiceImpl) DriverSummary(ctx *gin.Context) {
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

	driverUUID := profile.UUID // ✅ pastikan profile punya user uuid
	if driverUUID == "" {
		// fallback kalau struktur profile kamu beda:
		driverUUID = profile.UUID
	}

	if driverUUID == "" {
		helpers.JsonErr[any](ctx, "driver uuid not found", http.StatusUnauthorized, errors.New("missing driver_uuid"))
		return
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	dayStart := startOfDay(now)
	dayEnd := endOfDay(now)

	jobToday, err := s.repo.CountDriverJobToday(driverUUID, dayStart, dayEnd)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count job today", http.StatusInternalServerError, err)
		return
	}

	jobWaiting, err := s.repo.CountDriverJobWaiting(driverUUID)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count job waiting", http.StatusInternalServerError, err)
		return
	}

	resp := dto.DriverDashboardSummary{
		JobToday:   jobToday,
		JobWaiting: jobWaiting,
	}

	helpers.JsonOK(ctx, "success", resp)
}
