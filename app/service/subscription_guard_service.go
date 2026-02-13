package service

import (
	"github.com/gin-gonic/gin"
	"harjonan.id/user-service/app/repository"
)

type SubscriptionGuardService interface {
	CheckClientAccess(ctx *gin.Context, clientUUID string) (bool, any, error)
}

type SubscriptionGuardServiceImpl struct {
	repo repository.ClientSubscriptionRepository
}

func NewSubscriptionGuardService(repo repository.ClientSubscriptionRepository) *SubscriptionGuardServiceImpl {
	return &SubscriptionGuardServiceImpl{repo: repo}
}

func (s *SubscriptionGuardServiceImpl) CheckClientAccess(ctx *gin.Context, clientUUID string) (bool, any, error) {
	allowed, sub, err := s.repo.IsClientAllowed(clientUUID)
	if err != nil {
		return false, nil, err
	}
	return allowed, sub, nil
}
