package handler

import (
	"encoding/json"
	"net/http"
	"subtracker/internal/domain/dto"
	"subtracker/internal/mapper"
	"subtracker/internal/service"
	"subtracker/pkg/logger"
	"subtracker/pkg/response"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SubscriptionHandler struct {
	service service.SubscriptionServiceInterface
	logger  logger.Logger
}

func NewSubscriptionHandler(service service.SubscriptionServiceInterface, logger logger.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
		logger:  logger,
	}
}

func (s *SubscriptionHandler) handleError(w http.ResponseWriter, r *http.Request, err error, message string, code int) {
	if err != nil {
		s.logger.Error(message,
			zap.Error(err),
			zap.Int("code", code),
			zap.String("url", r.URL.Path),
		)
	} else {
		s.logger.Error(message,
			zap.Int("code", code),
			zap.String("url", r.URL.Path),
		)
	}

	jsonErr := response.APIError{
		Code:     code,
		Message:  message,
		Resource: r.URL.Path,
	}
	jsonErr.Send(w)
}

func (s *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var subscription dto.CreateSubscriptionRequest
	if err := decoder.Decode(&subscription); err != nil {
		s.logger.Error("Failed to decode request body", zap.Error(err))
		s.handleError(w, r, err, "Invalid request body", http.StatusBadRequest)
		return
	}
	if subscription.ServiceName == "" || subscription.Price <= 0 || subscription.UserID == "" || subscription.StartDate == "" {
		s.handleError(w, r, nil, "Missing required fields", http.StatusBadRequest)
		return
	}
	_, err := uuid.Parse(subscription.UserID)
	if err != nil {
		s.handleError(w, r, err, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	sub, err := mapper.ToDomainFromDTO(subscription)
	if err != nil {
		s.handleError(w, r, err, "Failed to map DTO to domain", http.StatusInternalServerError)
		return
	}
	if err := s.service.CreateSubscription(r.Context(), sub); err != nil {
		s.handleError(w, r, err, "Failed to create subscription", http.StatusInternalServerError)
		return
	}
}

func (s *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	// Implementation for listing subscriptions
}
