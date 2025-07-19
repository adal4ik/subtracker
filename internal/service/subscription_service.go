package service

import (
	"context"

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
	return subDomainList, nil
}

func (s *SubscriptionService) GetSubscription(ctx context.Context, id string) (domain.Subscription, error) {
	subDao, err := s.repo.GetSubscription(ctx, id)
	if err != nil {
		return domain.Subscription{}, err
	}
	return mapper.ToDomainFromDAO(subDao), nil
}

func (s *SubscriptionService) UpdateSubscription(ctx context.Context, subToUpdate domain.Subscription) error {
	s.logger.Debug("Attempting to update subscription", zap.String("id", subToUpdate.ID.String()))

	existingSubDAO, err := s.repo.GetSubscription(ctx, subToUpdate.ID.String())
	if err != nil {
		return err
	}

	finalSubDAO := dao.SubscriptionRow{
		ID:          existingSubDAO.ID,
		UserID:      existingSubDAO.UserID,
		ServiceName: subToUpdate.ServiceName,
		Price:       subToUpdate.Price,
		StartDate:   subToUpdate.StartDate,
		EndDate:     subToUpdate.EndDate,
	}

	return s.repo.UpdateSubscription(ctx, finalSubDAO)
}

func (s *SubscriptionService) DeleteSubscription(ctx context.Context, id string) error {
	s.logger.Debug("Deleting subscription", zap.String("id", id))
	return s.repo.DeleteSubscription(ctx, id)
}

func (s *SubscriptionService) CalculateCost(ctx context.Context, filter dto.CostFilter) (int, error) {
	s.logger.Debug("Calculating cost", zap.Any("filter", filter))

	subscriptions, err := s.repo.ListForCostCalculation(ctx, filter)
	if err != nil {
		return 0, err
	}

	totalCost := 0

	periodEndEffective := filter.PeriodEnd.AddDate(0, 1, -1)

	for _, sub := range subscriptions {
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
		if periodEndEffective.Before(overlapEnd) {
			overlapEnd = periodEndEffective
		}

		if !overlapStart.After(overlapEnd) {
			months := (overlapEnd.Year()-overlapStart.Year())*12 + int(overlapEnd.Month()) - int(overlapStart.Month()) + 1
			totalCost += sub.Price * months
		}
	}

	s.logger.Info("Cost calculated successfully", zap.Int("total_cost", totalCost))
	return totalCost, nil
}
