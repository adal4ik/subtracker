package service

import (
	"context"
	"subtracker/internal/domain"
	"subtracker/internal/mapper"
	"subtracker/internal/repository"
	"subtracker/pkg/logger"

	"go.uber.org/zap"
)

type SubscriptionServiceInterface interface {
	CreateSubscription(ctx context.Context, subDomain domain.Subscription) error
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

func (s *SubscriptionService) CreateSubscription(ctx context.Context, subDomain domain.Subscription) error {
	s.logger.Debug("Creating subscription", zap.String("service_name", subDomain.ServiceName),
		zap.Int("price", subDomain.Price),
		zap.String("user_id", subDomain.UserID.String()),
		zap.Time("start_date", subDomain.StartDate),
		zap.Any("end_date", subDomain.EndDate),
	)
	subDao := mapper.ToDAOFromDomain(subDomain)
	if err := s.repo.CreateSubscription(ctx, subDao); err != nil {
		s.logger.Error("Failed to create subscription", zap.Error(err),
			zap.String("service_name", subDomain.ServiceName),
			zap.Int("price", subDomain.Price),
			zap.String("user_id", subDomain.UserID.String()),
		)
		return err
	}
	s.logger.Info("Subscription created successfully", zap.String("id", subDomain.ID.String()))
	return nil
}
