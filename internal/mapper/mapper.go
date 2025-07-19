package mapper

import (
	"time"

	"subtracker/internal/domain"
	"subtracker/internal/domain/dao"
	"subtracker/internal/domain/dto"

	"github.com/google/uuid"
)

// DTO -> DOMAIN
func ToDomainFromDTO(req dto.CreateSubscriptionRequest) (domain.Subscription, error) {
	start, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		return domain.Subscription{}, err
	}

	var end *time.Time
	if req.EndDate != "" {
		t, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			return domain.Subscription{}, err
		}
		end = &t
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return domain.Subscription{}, err
	}

	return domain.Subscription{
		UserID:      userID,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		StartDate:   start,
		EndDate:     end,
	}, nil
}

// DOMAIN -> DTO
func ToDTOFromDomain(sub domain.Subscription) dto.SubscriptionResponse {
	start := sub.StartDate.Format("01-2006")

	var end string
	if sub.EndDate != nil {
		end = sub.EndDate.Format("01-2006")
	}

	return dto.SubscriptionResponse{
		ID:          sub.ID.String(),
		UserID:      sub.UserID.String(),
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		StartDate:   start,
		EndDate:     end,
	}
}

// DAO -> DOMAIN
func ToDomainFromDAO(row dao.SubscriptionRow) domain.Subscription {
	return domain.Subscription{
		ID:          row.ID,
		UserID:      row.UserID,
		ServiceName: row.ServiceName,
		Price:       row.Price,
		StartDate:   row.StartDate,
		EndDate:     row.EndDate,
	}
}

// DOMAIN -> DAO
func ToDAOFromDomain(sub domain.Subscription) dao.SubscriptionRow {
	return dao.SubscriptionRow{
		ID:          sub.ID,
		UserID:      sub.UserID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		StartDate:   sub.StartDate,
		EndDate:     sub.EndDate,
	}
}

func ToDomainFromUpdateDTO(req dto.UpdateSubscriptionRequest) (domain.Subscription, error) {
	start, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		return domain.Subscription{}, err
	}

	var end *time.Time
	if req.EndDate != "" {
		t, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			return domain.Subscription{}, err
		}
		end = &t
	}

	return domain.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		StartDate:   start,
		EndDate:     end,
	}, nil
}
