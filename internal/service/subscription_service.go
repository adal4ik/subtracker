package service

import (
	"context"
	"subtracker/internal/domain"
	"subtracker/internal/domain/dto"
	"subtracker/internal/mapper"
	"subtracker/internal/repository"
	"subtracker/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SubscriptionServiceInterface interface {
	CreateSubscription(ctx context.Context, subDomain domain.Subscription) error
	ListSubscriptions(ctx context.Context, filter dto.SubscriptionFilter) ([]domain.Subscription, error)
	GetSubscription(ctx context.Context, id string) (domain.Subscription, error)
	UpdateSubscription(ctx context.Context, subDomain domain.Subscription) error
	DeleteSubscription(ctx context.Context, id string) error
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
	if subDomain.ID == uuid.Nil {
		subDomain.ID = uuid.New()
	}
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

func (s *SubscriptionService) ListSubscriptions(ctx context.Context, filter dto.SubscriptionFilter) ([]domain.Subscription, error) {
	s.logger.Debug("Filtering subscriptions", zap.String("user_id", filter.UserID),
		zap.String("service_name", filter.ServiceName),
		zap.String("start_date", filter.StartDate),
		zap.String("end_date", filter.EndDate),
		zap.Int("limit", filter.Limit),
		zap.Int("offset", filter.Offset),
	)
	subscriptions, err := s.repo.ListSubscriptions(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to list subscriptions", zap.Error(err), zap.String("user_id", filter.UserID))
		return nil, err
	}
	subDomainList := make([]domain.Subscription, len(subscriptions))
	for i, sub := range subscriptions {
		subDomainList[i] = mapper.ToDomainFromDAO(sub)
	}
	s.logger.Info("Subscriptions listed successfully", zap.Int("count", len(subDomainList)))
	return subDomainList, nil
}

func (s *SubscriptionService) GetSubscription(ctx context.Context, id string) (domain.Subscription, error) {
	// Implementation for getting a specific subscription
	return domain.Subscription{}, nil
}
func (s *SubscriptionService) UpdateSubscription(ctx context.Context, subDomain domain.Subscription) error {
	// Implementation for updating a subscription
	return nil
}
func (s *SubscriptionService) DeleteSubscription(ctx context.Context, id string) error {
	// Implementation for deleting a subscription
	return nil
}
