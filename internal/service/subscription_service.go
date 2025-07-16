package service

import (
	"subtracker/internal/repository"
	"subtracker/pkg/logger"
)

type SubscriptionServiceInterface interface {
	// Define methods that SubscriptionService should implement
}

type SubscriptionService struct {
	repo   repository.SubscriptionRepositoryInterface
	logger logger.Logger
}

func NewSubscriptionService(repo repository.SubscriptionRepositoryInterface, logger logger.Logger) *SubscriptionService {
	return &SubscriptionService{
		repo:   repo,
		logger: logger,
	}
}
