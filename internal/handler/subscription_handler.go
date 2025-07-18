package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"subtracker/internal/domain/dto"
	"subtracker/internal/mapper"
	"subtracker/internal/service"
	"subtracker/pkg/apperrors"
	"subtracker/pkg/logger"
	"subtracker/pkg/response"
	"subtracker/utils"

	"github.com/go-chi/chi/v5"
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

func (s *SubscriptionHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	s.logger.Error("API Error",
		zap.Error(err),
		zap.String("url", r.URL.Path),
	)

	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		jsonErr := response.APIError{
			Code:     appErr.Code,
			Message:  appErr.Message,
			Resource: r.URL.Path,
		}
		jsonErr.Send(w)
		return
	}

	jsonErr := response.APIError{
		Code:     http.StatusInternalServerError,
		Message:  "Internal Server Error",
		Resource: r.URL.Path,
	}
	jsonErr.Send(w)
}

func (s *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.handleError(w, r, apperrors.NewBadRequest("invalid request body", err))
		return
	}
	if req.ServiceName == "" || req.Price < 0 || req.UserID == "" || req.StartDate == "" {
		s.handleError(w, r, apperrors.NewBadRequest("missing required fields", nil))
		return
	}
	if _, err := uuid.Parse(req.UserID); err != nil {
		s.handleError(w, r, apperrors.NewBadRequest("invalid user ID format", err))
		return
	}

	sub, err := mapper.ToDomainFromDTO(req)
	if err != nil {
		s.handleError(w, r, apperrors.NewBadRequest("failed to parse date", err))
		return
	}

	if err := s.service.CreateSubscription(r.Context(), sub); err != nil {
		s.handleError(w, r, err)
		return
	}

	response.APIResponse{Code: http.StatusCreated, Message: "Subscription created successfully"}.Send(w)
}

func (s *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := dto.SubscriptionFilter{
		UserID:      query.Get("user_id"),
		ServiceName: query.Get("service_name"),
		StartDate:   query.Get("start_date"),
		EndDate:     query.Get("end_date"),
		MinPrice:    utils.ParseFloatOrDefault(query.Get("min_price"), 0),
		MaxPrice:    utils.ParseFloatOrDefault(query.Get("max_price"), 0),
		HasEndDate:  utils.ParseBoolPointer(query.Get("has_end_date")),
		Limit:       utils.ParseIntOrDefault(query.Get("limit"), 10),
		Offset:      utils.ParseIntOrDefault(query.Get("offset"), 0),
	}

	result, err := s.service.ListSubscriptions(r.Context(), filter)
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	responseDTOs := make([]dto.SubscriptionResponse, len(result))
	for i, sub := range result {
		responseDTOs[i] = mapper.ToDTOFromDomain(sub)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseDTOs)
}

func (s *SubscriptionHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		s.handleError(w, r, apperrors.NewBadRequest("invalid subscription ID format", err))
		return
	}

	subscription, err := s.service.GetSubscription(r.Context(), id)
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapper.ToDTOFromDomain(subscription))
}

func (s *SubscriptionHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	// Implementation for updating a subscription
}
func (s *SubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	// Implementation for deleting a subscription
}
