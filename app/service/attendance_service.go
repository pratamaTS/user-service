package service

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"harjonan.id/user-service/app/domain/dto"
	"harjonan.id/user-service/app/helpers"
	"harjonan.id/user-service/app/repository"
)

type AttendanceService interface {
	Upsert(ctx *gin.Context)
	List(ctx *gin.Context)
}

type AttendanceServiceImpl struct {
	repo repository.AttendanceRepository
}

func NewAttendanceService(repo repository.AttendanceRepository) *AttendanceServiceImpl {
	return &AttendanceServiceImpl{repo: repo}
}

func (s *AttendanceServiceImpl) Upsert(ctx *gin.Context) {
	var req dto.AttendanceUpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}

	// ✅ ambil dari JWT middleware kalau ada
	if v, ok := ctx.Get("user_uuid"); ok {
		if sv, ok := v.(string); ok && sv != "" {
			req.UserUUID = sv
		}
	}
	if v, ok := ctx.Get("user_name"); ok {
		if sv, ok := v.(string); ok && sv != "" {
			req.UserName = sv
		}
	}

	if req.UserUUID == "" {
		helpers.JsonErr[any](ctx, "missing identifier", http.StatusBadRequest, errors.New("user_uuid required"))
		return
	}
	if req.BranchUUID == "" {
		helpers.JsonErr[any](ctx, "missing identifier", http.StatusBadRequest, errors.New("branch_uuid required"))
		return
	}
	if req.LocationLatitude == 0 && req.LocationLongitude == 0 {
		helpers.JsonErr[any](ctx, "missing identifier", http.StatusBadRequest, errors.New("location required"))
		return
	}

	res, err := s.repo.UpsertAttendance(&req)
	if err != nil {
		// rules error -> 400 biar enak ditangani di frontend (radius, full staff, dll)
		helpers.JsonErr[any](ctx, err.Error(), http.StatusBadRequest, err)
		return
	}
	helpers.JsonOK(ctx, "success", res)
}

func (s *AttendanceServiceImpl) List(ctx *gin.Context) {
	var req dto.FilterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	data, err := s.repo.ListAttendance(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to list attendance", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", data)
}
