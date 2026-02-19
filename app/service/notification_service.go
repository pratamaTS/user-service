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

type NotificationService interface {
	Fetch(ctx *gin.Context)
	MarkRead(ctx *gin.Context)
	MarkReadAll(ctx *gin.Context)
	Clear(ctx *gin.Context)
	ClearAll(ctx *gin.Context)
}

type NotificationServiceImpl struct {
	repo     repository.NotificationRepository
	authRepo repository.AuthRepository
}

func NewNotificationService(repo repository.NotificationRepository, authRepo repository.AuthRepository) *NotificationServiceImpl {
	return &NotificationServiceImpl{repo: repo, authRepo: authRepo}
}

// =========================
// POST /notifications/fetch
// Body: dto.FilterRequest
// - Owner: filter client_uuid saja (lihat semua)
// - Non-owner: filter client_uuid + (branch_uuid user OR user_uuid user)
// =========================
func (s *NotificationServiceImpl) Fetch(ctx *gin.Context) {
	var req dto.FilterRequest
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

	userUUID := profile.UUID
	branchUUID := profile.Branch.UUID
	roleValue := profile.Role.Value

	// defaults
	if req.SortBy == nil {
		req.SortBy = map[string]any{"created_at": "desc"}
	}
	if req.FilterBy == nil {
		req.FilterBy = map[string]any{}
	}

	// ✅ wajib scope by client
	req.FilterBy["client_uuid"] = clientUUID

	isOwner := roleValue == "OWNER" || roleValue == "SUPERADMIN"

	if !isOwner {
		// Non-owner: lihat notif untuk branch dia, atau notif personal (driver)
		req.FilterBy["$or"] = []map[string]any{
			{"branch_uuid": branchUUID},
			{"user_uuid": userUUID},
		}
	}

	list, total, err := s.repo.List(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to fetch notifications", http.StatusInternalServerError, err)
		return
	}

	resp := map[string]any{
		"items": list,
		"pagination": map[string]any{
			"page":      req.Pagination.Page,
			"page_size": req.Pagination.PageSize,
			"total":     total,
		},
	}

	helpers.JsonOK(ctx, "success", resp)
}

// =========================
// POST /notifications/:uuid/read
// =========================
func (s *NotificationServiceImpl) MarkRead(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "uuid required", http.StatusBadRequest, errors.New("uuid required"))
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

	clientUUID := profile.Client.UUID
	if clientUUID == "" {
		helpers.JsonErr[any](ctx, "client uuid not found", http.StatusUnauthorized, errors.New("missing client_uuid"))
		return
	}

	if err := s.repo.MarkRead(uuid, clientUUID); err != nil {
		helpers.JsonErr[any](ctx, "failed to mark read", http.StatusInternalServerError, err)
		return
	}

	helpers.JsonOK(ctx, "success", true)
}

// =========================
// POST /notifications/read-all
// Owner: mark all in client
// Non-owner: mark all in branch (dan bisa juga mark personal by user_uuid? -> kita keep simple branch)
// =========================
func (s *NotificationServiceImpl) MarkReadAll(ctx *gin.Context) {
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
	branchUUID := profile.Branch.UUID
	roleValue := profile.Role.Value

	if clientUUID == "" {
		helpers.JsonErr[any](ctx, "client uuid not found", http.StatusUnauthorized, errors.New("missing client_uuid"))
		return
	}

	isOwner := roleValue == "OWNER" || roleValue == "SUPERADMIN"

	targetBranch := branchUUID
	if isOwner {
		targetBranch = "" // ✅ mark all for client
	} else if targetBranch == "" {
		helpers.JsonErr[any](ctx, "branch uuid not found", http.StatusBadRequest, errors.New("missing branch_uuid"))
		return
	}

	updated, err := s.repo.MarkReadAll(clientUUID, targetBranch)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to mark read all", http.StatusInternalServerError, err)
		return
	}

	helpers.JsonOK(ctx, "success", map[string]any{"updated": updated, "ts": time.Now().Format(time.RFC3339)})
}

// =========================
// DELETE /notifications/clear
// Owner: clear all by client
// Non-owner: clear by branch
// =========================
func (s *NotificationServiceImpl) Clear(ctx *gin.Context) {
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
	branchUUID := profile.Branch.UUID
	roleValue := profile.Role.Value

	if clientUUID == "" {
		helpers.JsonErr[any](ctx, "client uuid not found", http.StatusUnauthorized, errors.New("missing client_uuid"))
		return
	}

	isOwner := roleValue == "OWNER" || roleValue == "SUPERADMIN"

	targetBranch := branchUUID
	if isOwner {
		targetBranch = "" // ✅ clear all for client
	} else if targetBranch == "" {
		helpers.JsonErr[any](ctx, "branch uuid not found", http.StatusBadRequest, errors.New("missing branch_uuid"))
		return
	}

	deleted, err := s.repo.Clear(clientUUID, targetBranch)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to clear", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", map[string]any{"deleted": deleted})
}

// =========================
// DELETE /notifications/clear-all
// (explicit) clear all by client (Owner only recommended)
// =========================
func (s *NotificationServiceImpl) ClearAll(ctx *gin.Context) {
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

	deleted, err := s.repo.ClearAll(clientUUID)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to clear all", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", map[string]any{"deleted": deleted})
}
