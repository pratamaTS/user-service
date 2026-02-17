package service

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"harjonan.id/user-service/app/domain/dto"
	"harjonan.id/user-service/app/helpers"
	"harjonan.id/user-service/app/repository"
)

type DashboardService interface {
	OwnerSummary(ctx *gin.Context)
}

type DashboardServiceImpl struct {
	repo     repository.DashboardRepository
	authRepo repository.AuthRepository
}

func NewDashboardService(repo repository.DashboardRepository, authRepo repository.AuthRepository) *DashboardServiceImpl {
	return &DashboardServiceImpl{repo: repo, authRepo: authRepo}
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	return startOfDay(t).Add(24 * time.Hour)
}

func startOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

func endOfMonth(t time.Time) time.Time {
	return startOfMonth(t).AddDate(0, 1, 0)
}

func (s *DashboardServiceImpl) OwnerSummary(ctx *gin.Context) {
	// request optional
	var req dto.OwnerDashboardRequest
	_ = ctx.ShouldBindJSON(&req)

	// ✅ ambil client_uuid dari JWT middleware kamu
	// pastikan middleware set ctx.Set("client_uuid", ...)
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

	// threshold low stock default
	threshold := req.LowStockThreshold
	if threshold <= 0 {
		threshold = 10
	}

	// timezone: WIB
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	dayStart := startOfDay(now)
	dayEnd := endOfDay(now)
	monthStart := startOfMonth(now)
	monthEnd := endOfMonth(now)

	// ✅ ambil semua branch UUID untuk owner/company scope
	branchUUIDs, err := s.repo.GetCompanyBranchUUIDs(clientUUID)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to get branches", http.StatusInternalServerError, err)
		return
	}

	// ✅ Total Product (repo kamu sudah force ke branch gudang)
	totalProduct, err := s.repo.CountActiveProducts(clientUUID)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count products", http.StatusInternalServerError, err)
		return
	}

	// ✅ Total Stock Request (month)
	totalStockReq, err := s.repo.CountStockRequestMonth(branchUUIDs, monthStart, monthEnd)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count stock request", http.StatusInternalServerError, err)
		return
	}

	// ✅ Stock Request Process
	stockReqProcess, err := s.repo.CountStockRequestProcess(branchUUIDs)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count stock request process", http.StatusInternalServerError, err)
		return
	}

	// ✅ Job Waiting Accept
	jobWaitingAccept, err := s.repo.CountJobWaitingAccept(branchUUIDs)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count job waiting accept", http.StatusInternalServerError, err)
		return
	}

	// ✅ Transaction Today
	txToday, err := s.repo.CountTransactionToday(branchUUIDs, dayStart, dayEnd)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count transaction today", http.StatusInternalServerError, err)
		return
	}

	// ✅ Revenue Month
	revenueMonth, err := s.repo.SumRevenueMonth(branchUUIDs, monthStart, monthEnd)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to sum revenue month", http.StatusInternalServerError, err)
		return
	}

	// ✅ Low Stock SKU (all branches)
	lowStockSKU, err := s.repo.CountLowStockSKU(branchUUIDs, threshold)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count low stock sku", http.StatusInternalServerError, err)
		return
	}

	// ✅ Driver Available (sesuai repo kamu -> hitung jumlah job in progress untuk driver)
	driverAvailable, err := s.repo.CountDriverAvailable(clientUUID)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to count driver available", http.StatusInternalServerError, err)
		return
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
