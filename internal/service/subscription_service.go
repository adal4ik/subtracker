package service

import (
	"context"
	"time"

	"subtracker/internal/domain"
	"subtracker/internal/domain/dao"
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
	CalculateCost(ctx context.Context, filter dto.CostFilter) (int, error)
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
	s.logger.Debug("Entering CreateSubscription service",
		zap.String("service_name", subDomain.ServiceName),
		zap.String("user_id", subDomain.UserID.String()),
	)
	if subDomain.ID == uuid.Nil {
		subDomain.ID = uuid.New()
		s.logger.Debug("Generated new subscription ID", zap.String("subscription_id", subDomain.ID.String()))
	}
	subDao := mapper.ToDAOFromDomain(subDomain)
	return s.repo.CreateSubscription(ctx, subDao)
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
		return nil, err
	}
	subDomainList := make([]domain.Subscription, len(subscriptions))
	for i, sub := range subscriptions {
		subDomainList[i] = mapper.ToDomainFromDAO(sub)
	}
	s.logger.Debug("Exiting ListSubscriptions service", zap.Int("count", len(subDomainList)))

	return subDomainList, nil
}

func (s *SubscriptionService) GetSubscription(ctx context.Context, id string) (domain.Subscription, error) {
	s.logger.Debug("Entering GetSubscription service", zap.String("id", id))
	subDao, err := s.repo.GetSubscription(ctx, id)
	if err != nil {
		return domain.Subscription{}, err
	}
	return mapper.ToDomainFromDAO(subDao), nil
}

func (s *SubscriptionService) UpdateSubscription(ctx context.Context, subToUpdate domain.Subscription) error {
	s.logger.Debug("Entering UpdateSubscription service",
		zap.String("subscription_id", subToUpdate.ID.String()),
		zap.Any("updates", subToUpdate),
	)

	existingSubDAO, err := s.repo.GetSubscription(ctx, subToUpdate.ID.String())
	if err != nil {
		return err
	}

	s.logger.Debug("Found existing subscription to update", zap.Any("existing_dao", existingSubDAO))

	finalSubDAO := dao.SubscriptionRow{
		ID:          existingSubDAO.ID,
		UserID:      existingSubDAO.UserID,
		ServiceName: subToUpdate.ServiceName,
		Price:       subToUpdate.Price,
		StartDate:   subToUpdate.StartDate,
		EndDate:     subToUpdate.EndDate,
	}

	s.logger.Debug("Proceeding to update with final DAO object", zap.Any("final_dao", finalSubDAO))

	return s.repo.UpdateSubscription(ctx, finalSubDAO)
}

func (s *SubscriptionService) DeleteSubscription(ctx context.Context, id string) error {
	s.logger.Debug("Entering DeleteSubscription service", zap.String("id", id))

	err := s.repo.DeleteSubscription(ctx, id)
	if err != nil {
		return err
	}

	s.logger.Debug("Exiting DeleteSubscription service", zap.String("id", id))
	return nil
}

func (s *SubscriptionService) CalculateCost(ctx context.Context, filter dto.CostFilter) (int, error) {
	s.logger.Debug("Entering CalculateCost service", zap.Any("filter", filter))

	subscriptions, err := s.repo.ListForCostCalculation(ctx, filter)
	if err != nil {
		return 0, err
	}

	s.logger.Debug("Found subscriptions for calculation", zap.Int("count", len(subscriptions)))

	totalCost := 0
	periodEndEffective := filter.PeriodEnd.AddDate(0, 1, 0).Add(-1 * time.Nanosecond)

	for _, sub := range subscriptions {
		s.logger.Debug("Processing subscription for cost calculation",
			zap.String("subscription_id", sub.ID.String()),
			zap.Time("sub_start_date", sub.StartDate),
			zap.Any("sub_end_date", sub.EndDate),
			zap.Int("sub_price", sub.Price),
		)

		subStart := sub.StartDate
		subEnd := periodEndEffective
		if sub.EndDate != nil && sub.EndDate.Before(periodEndEffective) {
			subEnd = *sub.EndDate
		}

		overlapStart := filter.PeriodStart
		if subStart.After(overlapStart) {
			overlapStart = subStart
		}

		overlapEnd := subEnd

		if overlapStart.After(overlapEnd) {
			s.logger.Debug("Subscription is outside the calculation period, skipping.", zap.String("subscription_id", sub.ID.String()))
			continue
		}

		months := (overlapEnd.Year()-overlapStart.Year())*12 + int(overlapEnd.Month()) - int(overlapStart.Month()) + 1
		costForSub := sub.Price * months
		totalCost += costForSub

		s.logger.Debug("Calculated cost for one subscription",
			zap.String("subscription_id", sub.ID.String()),
			zap.Time("overlap_start", overlapStart),
			zap.Time("overlap_end", overlapEnd),
			zap.Int("months_counted", months),
			zap.Int("cost_for_this_sub", costForSub),
		)
	}

	s.logger.Info("Total cost calculated successfully", zap.Int("total_cost", totalCost))
	return totalCost, nil
}
