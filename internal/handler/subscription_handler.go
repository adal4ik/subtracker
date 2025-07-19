package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"subtracker/internal/domain/dto"
	"subtracker/internal/mapper"
	"subtracker/internal/service"
	"subtracker/pkg/apperrors"
	"subtracker/pkg/logger"
	"subtracker/pkg/response"
	"subtracker/pkg/validator"
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

// @Summary      Create Subscription
// @Description  Adds a new subscription to the system based on the provided data.
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Param        subscription body dto.CreateSubscriptionRequest true "Subscription Information"
// @Success      201  {object}  response.APIResponse
// @Failure      400  {object}  apperrors.AppError "Invalid request body or fields"
// @Failure      409  {object}  apperrors.AppError "Conflict if subscription with this ID already exists"
// @Failure      500  {object}  apperrors.AppError "Internal server error"
// @Router       /subscriptions [post]
func (s *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.handleError(w, r, apperrors.NewBadRequest("invalid request body", err))
		return
	}
	if err := validator.ValidateStruct(req); err != nil {
		s.handleError(w, r, apperrors.NewBadRequest("validation failed", err))
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

// @Summary      List Subscriptions
// @Description  Gets a list of subscriptions with filtering and pagination.
// @Tags         Subscriptions
// @Produce      json
// @Param        user_id      query     string  false  "Filter by User ID (UUID)"
// @Param        service_name query     string  false  "Filter by Service Name"
// @Param        min_price    query     int     false  "Filter by minimum price"
// @Param        max_price    query     int     false  "Filter by maximum price"
// @Param        start_date   query     string  false  "Filter by start date (format: MM-YYYY)"
// @Param        end_date     query     string  false  "Filter by end date (format: MM-YYYY)"
// @Param        has_end_date query     bool    false  "Filter by presence of an end date"
// @Param        limit        query     int     false  "Pagination limit (default 10, max 100)"
// @Param        offset       query     int     false  "Pagination offset (default 0)"
// @Success      200  {array}   dto.SubscriptionResponse
// @Failure      400  {object}  apperrors.AppError "Invalid filter parameters"
// @Failure      500  {object}  apperrors.AppError "Internal server error"
// @Router       /subscriptions [get]
func (s *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := dto.SubscriptionFilter{
		UserID:      query.Get("user_id"),
		ServiceName: query.Get("service_name"),
		StartDate:   query.Get("start_date"),
		EndDate:     query.Get("end_date"),
		MinPrice:    utils.ParseIntOrDefault(query.Get("min_price"), 0),
		MaxPrice:    utils.ParseIntOrDefault(query.Get("max_price"), 0),
		HasEndDate:  utils.ParseBoolPointer(query.Get("has_end_date")),
		Limit:       utils.ParseIntOrDefault(query.Get("limit"), 10),
		Offset:      utils.ParseIntOrDefault(query.Get("offset"), 0),
	}
	if err := validator.ValidateStruct(filter); err != nil {
		s.handleError(w, r, apperrors.NewBadRequest("invalid filter parameters", err))
		return
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

// @Summary      Get Subscription by ID
// @Description  Retrieves a single subscription by its unique ID.
// @Tags         Subscriptions
// @Produce      json
// @Param        id   path      string  true  "Subscription ID (UUID format)"
// @Success      200  {object}  dto.SubscriptionResponse
// @Failure      400  {object}  apperrors.AppError "Invalid ID format"
// @Failure      404  {object}  apperrors.AppError "Subscription not found"
// @Failure      500  {object}  apperrors.AppError "Internal server error"
// @Router       /subscriptions/{id} [get]
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

// @Summary      Update Subscription
// @Description  Updates an existing subscription's details by its ID. UserID cannot be changed.
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Param        id           path      string                       true  "Subscription ID (UUID format)"
// @Param        subscription body      dto.UpdateSubscriptionRequest true  "Fields to update"
// @Success      200          {object}  response.APIResponse
// @Failure      400          {object}  apperrors.AppError "Invalid ID format or request body"
// @Failure      404          {object}  apperrors.AppError "Subscription not found"
// @Failure      500          {object}  apperrors.AppError "Internal server error"
// @Router       /subscriptions/{id} [put]
func (s *SubscriptionHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		s.handleError(w, r, apperrors.NewBadRequest("invalid subscription ID format", err))
		return
	}

	var req dto.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.handleError(w, r, apperrors.NewBadRequest("invalid request body", err))
		return
	}

	if err := validator.ValidateStruct(req); err != nil {
		s.handleError(w, r, apperrors.NewBadRequest("validation failed", err))
		return
	}

	sub, err := mapper.ToDomainFromUpdateDTO(req)
	if err != nil {
		s.handleError(w, r, apperrors.NewBadRequest("failed to parse date", err))
		return
	}

	sub.ID = id

	if err := s.service.UpdateSubscription(r.Context(), sub); err != nil {
		s.handleError(w, r, err)
		return
	}

	response.APIResponse{Code: http.StatusOK, Message: "Subscription updated successfully"}.Send(w)
}

// @Summary      Delete Subscription
// @Description  Deletes a subscription by its unique ID.
// @Tags         Subscriptions
// @Produce      json
// @Param        id   path      string  true  "Subscription ID (UUID format)"
// @Success      204  "No Content"
// @Failure      400  {object}  apperrors.AppError "Invalid ID format"
// @Failure      404  {object}  apperrors.AppError "Subscription not found"
// @Failure      500  {object}  apperrors.AppError "Internal server error"
// @Router       /subscriptions/{id} [delete]
func (s *SubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		s.handleError(w, r, apperrors.NewBadRequest("invalid subscription ID format", err))
		return
	}

	if err := s.service.DeleteSubscription(r.Context(), id); err != nil {
		s.handleError(w, r, err)
		return
	}

	response.APIResponse{Code: http.StatusNoContent, Message: "Subscription deleted successfully"}.Send(w)
}

// @Summary      Calculate Total Cost
// @Description  Calculates the total cost of subscriptions for a user over a specified period.
// @Tags         Subscriptions
// @Produce      json
// @Param        user_id      query     string  true   "User ID (UUID format) for whom to calculate the cost"
// @Param        period_start query     string  true   "Start of the calculation period (format: MM-YYYY)"
// @Param        period_end   query     string  true   "End of the calculation period (format: MM-YYYY)"
// @Param        service_name query     string  false  "Optional: filter by a specific service name"
// @Success      200          {object}  dto.CostResponse
// @Failure      400          {object}  apperrors.AppError "Invalid or missing parameters"
// @Failure      500          {object}  apperrors.AppError "Internal server error"
// @Router       /subscriptions/cost [get]
func (s *SubscriptionHandler) CalculateCost(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	costRequest := dto.CostRequest{
		UserID:      query.Get("user_id"),
		ServiceName: query.Get("service_name"),
		PeriodStart: query.Get("period_start"),
		PeriodEnd:   query.Get("period_end"),
	}

	if err := validator.ValidateStruct(costRequest); err != nil {
		s.handleError(w, r, apperrors.NewBadRequest("invalid query parameters", err))
		return
	}
	periodStart, _ := time.Parse("01-2006", costRequest.PeriodStart)
	periodEnd, _ := time.Parse("01-2006", costRequest.PeriodEnd)

	if periodEnd.Before(periodStart) {
		s.handleError(w, r, apperrors.NewBadRequest("period_end cannot be before period_start", nil))
		return
	}
	filter := dto.CostFilter{
		UserID:      costRequest.UserID,
		ServiceName: costRequest.ServiceName,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	}
	totalCost, err := s.service.CalculateCost(r.Context(), filter)
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	responseDTO := dto.CostResponse{TotalCost: totalCost}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseDTO)
}

func (s *SubscriptionHandler) ServeSwaggerJSON(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./docs/swagger.json")
}
