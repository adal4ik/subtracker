package service

import (
	"subtracker/internal/repository"
	"subtracker/pkg/logger"
)

type Service struct {
	SubscriptionService *SubscriptionService
}

func NewService(repo *repository.Repository, logger logger.Logger) *Service {
	return &Service{
		SubscriptionService: NewSubscriptionService(repo.SubscriptionRepository, logger),
	}
}
