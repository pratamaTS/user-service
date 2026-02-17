package service

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"harjonan.id/user-service/app/domain/dao"
	"harjonan.id/user-service/app/domain/dto"
	"harjonan.id/user-service/app/helpers"
	"harjonan.id/user-service/app/repository"
)

type AuthService interface {
	Register(ctx *gin.Context)
	Login(ctx *gin.Context)
	Refresh(ctx *gin.Context)
	Logout(ctx *gin.Context)
	Me(ctx *gin.Context)
}

type AuthServiceImpl struct {
	repo                 repository.AuthRepository
	companyRepo          repository.CompanyRepository
	roleRepo             repository.RoleRepository
	companyUserRepo      repository.CompanyUserRepository
	clientRepo           repository.ClientRepository
	clientUserRepo       repository.ClientUserRepository
	checkRateLimitRepo   repository.RateLimitRepository
	roleMenuAccessRepo   repository.RoleMenuAccessRepository
	subscriptionGuardSvc SubscriptionGuardService
}

func NewAuthService(
	repo repository.AuthRepository,
	companyRepo repository.CompanyRepository,
	roleRepo repository.RoleRepository,
	companyUserRepo repository.CompanyUserRepository,
	clientRepo repository.ClientRepository,
	clientUserRepo repository.ClientUserRepository,
	checkRateLimitRepo repository.RateLimitRepository,
	roleMenuAccessRepo repository.RoleMenuAccessRepository,
	subscriptionGuardSvc SubscriptionGuardService,
) *AuthServiceImpl {
	return &AuthServiceImpl{
		repo:                 repo,
		companyRepo:          companyRepo,
		roleRepo:             roleRepo,
		companyUserRepo:      companyUserRepo,
		clientRepo:           clientRepo,
		clientUserRepo:       clientUserRepo,
		checkRateLimitRepo:   checkRateLimitRepo,
		roleMenuAccessRepo:   roleMenuAccessRepo,
		subscriptionGuardSvc: subscriptionGuardSvc,
	}
}

func (s *AuthServiceImpl) Register(ctx *gin.Context) {
	var req dto.RequestParams[map[string]any]
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}

	uuid := getString(req.Params, "selected_uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing selected_uuid", http.StatusBadRequest, errors.New("selected_uuid required"))
		return
	}

	var subject string
	var role dao.Role

	ip := ctx.ClientIP()
	sessionID, _ := ctx.Get("session_id")

	sessionKey := fmt.Sprintf("register:session:%s", sessionID)
	ipKey := fmt.Sprintf("register:ip:%s", ip)

	if err := s.checkRateLimitRepo.CheckRateLimit(ctx, sessionKey, 5, 5*time.Minute); err != nil {
		helpers.JsonErr[any](ctx, "too many attempts for this session", http.StatusTooManyRequests, err)
		return
	}
	if err := s.checkRateLimitRepo.CheckRateLimit(ctx, ipKey, 5, 5*time.Minute); err != nil {
		helpers.JsonErr[any](ctx, "too many attempts from this IP", http.StatusTooManyRequests, err)
		return
	}

	result, err := s.companyRepo.DetailCompany(uuid)
	if err == nil || result.UUID != "" {
		subject = "COMPANY"
	} else {
		resultClient, err := s.clientRepo.DetailClient(uuid)
		if err != nil || resultClient.UUID == "" {
			helpers.JsonErr[any](ctx, "invalid company or client uuid", http.StatusBadRequest, errors.New("company or client not found"))
			return
		}
		subject = "CLIENT"
	}

	role, err = s.roleRepo.GetRoleByValue("OWNER", subject)
	if err != nil {
		helpers.JsonErr[any](ctx, "role not found", http.StatusInternalServerError, err)
		return
	}

	auth, err := s.repo.Register(req.Params, subject)
	if err != nil {
		helpers.JsonErr[any](ctx, "register failed", http.StatusInternalServerError, err)
		return
	}

	userUUID := auth.UserUUID
	if userUUID == "" {
		helpers.JsonErr[any](ctx, "missing user uuid", http.StatusInternalServerError, errors.New("user uuid not found in auth"))
		return
	}

	switch subject {
	case "COMPANY":
		if err == nil {
			_, _ = s.companyUserRepo.SaveCompanyUser(&dao.CompanyUser{
				CompanyUUID: uuid,
				RoleUUID:    role.UUID,
				AdminUUID:   userUUID,
			})
		}
	case "CLIENT":
		if err == nil {
			_, _ = s.clientUserRepo.SaveClientUser(&dao.ClientUser{
				ClientUUID: uuid,
				RoleUUID:   role.UUID,
				UserUUID:   userUUID,
			})
		}
	}

	helpers.JsonOK(ctx, "register success", map[string]any{
		"token":              auth.Token,
		"expired_at":         auth.ExpiredAt,
		"refresh_token":      auth.RefreshToken,
		"refresh_expired_at": auth.RefreshExpiredAt,
		"login_at":           auth.LoginAt,
		"status":             auth.Status,
		"subject":            auth.Subject,
		"session_id":         auth.SessionID,
		"user_uuid":          auth.UserUUID,
	})
}

func (s *AuthServiceImpl) Login(ctx *gin.Context) {
	var req dto.RequestParams[struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}]
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	if req.Params.Username == "" {
		helpers.JsonErr[any](ctx, "username required", http.StatusBadRequest, errors.New("missing username"))
		return
	}
	if req.Params.Password == "" {
		helpers.JsonErr[any](ctx, "password required", http.StatusBadRequest, errors.New("missing password"))
		return
	}

	ip := ctx.ClientIP()
	sessionID, _ := ctx.Get("session_id")

	sessionKey := fmt.Sprintf("login:session:%s", sessionID)
	ipKey := fmt.Sprintf("login:ip:%s", ip)

	if err := s.checkRateLimitRepo.CheckRateLimit(ctx, sessionKey, 5, 5*time.Minute); err != nil {
		helpers.JsonErr[any](ctx, "too many attempts for this session", http.StatusTooManyRequests, err)
		return
	}
	if err := s.checkRateLimitRepo.CheckRateLimit(ctx, ipKey, 5, 5*time.Minute); err != nil {
		helpers.JsonErr[any](ctx, "too many attempts from this IP", http.StatusTooManyRequests, err)
		return
	}

	auth, err := s.repo.Login(req.Params.Username, req.Params.Password)
	if err != nil {
		helpers.JsonErr[any](ctx, "login failed", http.StatusUnauthorized, err)
		return
	}

	payload := map[string]any{
		"token":              auth.Token,
		"expired_at":         auth.ExpiredAt,
		"refresh_token":      auth.RefreshToken,
		"refresh_expired_at": auth.RefreshExpiredAt,
		"login_at":           auth.LoginAt,
		"status":             auth.Status,
		"subject":            auth.Subject,
		"session_id":         auth.SessionID,
		"user_uuid":          auth.UserUUID,
	}

	helpers.JsonOK(ctx, "login success", payload)
}

func (s *AuthServiceImpl) Refresh(ctx *gin.Context) {
	var req dto.RequestParams[struct {
		RefreshToken string `json:"refresh_token"`
	}]
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}

	accessToken := ctx.GetString("access_token")
	if accessToken == "" {
		helpers.JsonErr[any](ctx, "missing access token", http.StatusBadRequest, errors.New("no bearer token"))
		return
	}

	auth, err := s.repo.Refresh(accessToken, req.Params.RefreshToken)
	if err != nil {
		helpers.JsonErr[any](ctx, "refresh failed", http.StatusUnauthorized, err)
		return
	}

	helpers.JsonOK(ctx, "refresh success", map[string]any{
		"token":              auth.Token,
		"expired_at":         auth.ExpiredAt,
		"refresh_token":      auth.RefreshToken,
		"refresh_expired_at": auth.RefreshExpiredAt,
		"login_at":           auth.LoginAt,
		"status":             auth.Status,
		"subject":            auth.Subject,
		"session_id":         auth.SessionID,
		"user_uuid":          auth.UserUUID,
	})
}

func (s *AuthServiceImpl) Logout(ctx *gin.Context) {
	accessToken := ctx.GetString("access_token")
	if accessToken == "" {
		helpers.JsonErr[any](ctx, "missing access token", http.StatusBadRequest, errors.New("no bearer token"))
		return
	}
	if err := s.repo.Logout(accessToken); err != nil {
		helpers.JsonErr[any](ctx, "logout failed", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK[any](ctx, "logout success", struct{}{})
}

func (s *AuthServiceImpl) Me(ctx *gin.Context) {
	accessToken := ctx.GetString("access_token")
	if accessToken == "" {
		helpers.JsonErr[any](ctx, "missing access token", http.StatusBadRequest, errors.New("no bearer token"))
		return
	}

	profile, err := s.repo.ValidateToken(accessToken)
	if err != nil {
		helpers.JsonErr[any](ctx, "validate token failed", http.StatusUnauthorized, err)
		return
	}

	var menus []dto.UserMenu

	if profile != nil && profile.Role.UUID != "" {
		menus, err = s.roleMenuAccessRepo.GetMenusByRole(ctx, profile.Role.UUID)
		if err != nil {
			helpers.JsonErr[any](ctx, "failed to get menus", http.StatusInternalServerError, err)
			return
		}
	}

	resp := map[string]any{
		"profile": profile,
		"menus":   menus,
	}

	if profile.Client.UUID != "" {
		resp["client_uuid"] = profile.Client.UUID

		if !profile.IsCompany {
			allowed, sub, subErr := s.subscriptionGuardSvc.CheckClientAccess(ctx, profile.Client.UUID)
			if sub != nil {
				resp["subscription"] = sub
			}

			if subErr != nil {
				helpers.JsonErr[any](ctx, "failed to check subscription", http.StatusInternalServerError, subErr)
				return
			}

			if !allowed {
				helpers.JsonErr[any](ctx, "subscription inactive or expired", http.StatusForbidden, nil)
				return
			}
		}
	}

	if profile.Role.Value != "" {
		resp["role_value"] = profile.Role.Value
	}

	helpers.JsonOK(ctx, "profile success", resp)
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
