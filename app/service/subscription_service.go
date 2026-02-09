package service

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"harjonan.id/user-service/app/domain/dao"
	"harjonan.id/user-service/app/domain/dto"
	"harjonan.id/user-service/app/helpers"
	"harjonan.id/user-service/app/repository"
)

type SubscriptionService interface {
	Upsert(ctx *gin.Context)
	Detail(ctx *gin.Context)
	List(ctx *gin.Context)
	Delete(ctx *gin.Context)

	ActivateFromMaster(ctx *gin.Context)
}

type SubscriptionServiceImpl struct {
	repo                   repository.SubscriptionRepository
	clientSubscriptionRepo repository.ClientSubscriptionRepository
}

func NewSubscriptionService(
	repo repository.SubscriptionRepository,
	clientSubscriptionRepo repository.ClientSubscriptionRepository,
) *SubscriptionServiceImpl {
	return &SubscriptionServiceImpl{
		repo:                   repo,
		clientSubscriptionRepo: clientSubscriptionRepo,
	}
}

func (s *SubscriptionServiceImpl) Upsert(ctx *gin.Context) {
	var req dao.Subscription
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}

	if req.Name == "" {
		helpers.JsonErr[any](ctx, "missing identifier", http.StatusBadRequest, errors.New("name required"))
		return
	}

	res, err := s.repo.Upsert(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to save subscription", http.StatusInternalServerError, err)
		return
	}

	helpers.JsonOK(ctx, "success", res)
}

func (s *SubscriptionServiceImpl) Detail(ctx *gin.Context) {
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

func (s *SubscriptionServiceImpl) List(ctx *gin.Context) {
	var req dto.APIRequest[dto.FilterRequest]
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}

	data, err := s.repo.List(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to list subscription", http.StatusInternalServerError, err)
		return
	}

	helpers.JsonOK(ctx, "success", data)
}

func (s *SubscriptionServiceImpl) Delete(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}

	if err := s.repo.Delete(uuid); err != nil {
		helpers.JsonErr[any](ctx, "failed to delete subscription", http.StatusInternalServerError, err)
		return
	}

	helpers.JsonOK[struct{}](ctx, "success", struct{}{})
}

func (s *SubscriptionServiceImpl) ActivateFromMaster(ctx *gin.Context) {
	var req struct {
		ClientUUID       string `json:"client_uuid"`
		SubscriptionUUID string `json:"subscription_uuid"`
		CreatedBy        string `json:"created_by"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}

	if req.ClientUUID == "" {
		helpers.JsonErr[any](ctx, "missing client uuid", http.StatusBadRequest, errors.New("client uuid required"))
		return
	}

	if req.SubscriptionUUID == "" {
		helpers.JsonErr[any](ctx, "missing subscription uuid", http.StatusBadRequest, errors.New("subscription uuid required"))
		return
	}

	res, err := s.clientSubscriptionRepo.ActivateFromMaster(req.ClientUUID, req.SubscriptionUUID, req.CreatedBy)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			helpers.JsonErr[any](ctx, "not found", http.StatusNotFound, err)
			return
		}
		helpers.JsonErr[any](ctx, "failed to activate subscription", http.StatusInternalServerError, err)
		return
	}

	helpers.JsonOK(ctx, "success", res)
}
