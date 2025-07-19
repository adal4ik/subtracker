package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"subtracker/internal/domain/dto"
	"subtracker/internal/mapper"
	"subtracker/internal/service"
	"subtracker/pkg/apperrors"
	"subtracker/pkg/logger"
	"subtracker/pkg/response"
	"subtracker/utils"
	"time"

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

func (s *SubscriptionHandler) validateListFilter(r *http.Request) error {
	query := r.URL.Query()

	if userIDStr := query.Get("user_id"); userIDStr != "" {
		if _, err := uuid.Parse(userIDStr); err != nil {
			return apperrors.NewBadRequest("invalid user_id format", err)
		}
	}

	if startDateStr := query.Get("start_date"); startDateStr != "" {
		if _, err := time.Parse("01-2006", startDateStr); err != nil {
			return apperrors.NewBadRequest("invalid start_date format, use MM-YYYY", err)
		}
	}

	if endDateStr := query.Get("end_date"); endDateStr != "" {
		if _, err := time.Parse("01-2006", endDateStr); err != nil {
			return apperrors.NewBadRequest("invalid end_date format, use MM-YYYY", err)
		}
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			return apperrors.NewBadRequest("limit must be a non-negative integer", err)
		}
		const maxLimit = 100
		if limit > maxLimit {
			return apperrors.NewBadRequest("limit exceeds maximum allowed value of 100", nil)
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err != nil || offset < 0 {
			return apperrors.NewBadRequest("offset must be a non-negative integer", err)
		}
	}

	var minPrice, maxPrice int
	var err error

	minPriceStr := query.Get("min_price")
	if minPriceStr != "" {
		minPrice, err = strconv.Atoi(minPriceStr)
		if err != nil || minPrice < 0 {
			return apperrors.NewBadRequest("min_price must be a non-negative integer", err)
		}
	}

	maxPriceStr := query.Get("max_price")
	if maxPriceStr != "" {
		maxPrice, err = strconv.Atoi(maxPriceStr)
		if err != nil || maxPrice < 0 {
			return apperrors.NewBadRequest("max_price must be a non-negative integer", err)
		}
	}

	if minPriceStr != "" && maxPriceStr != "" {
		if minPrice > maxPrice {
			return apperrors.NewBadRequest("min_price cannot be greater than max_price", nil)
		}
	}

	return nil
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
	if err := s.validateListFilter(r); err != nil {
		s.handleError(w, r, err)
		return
	}

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

	if req.ServiceName == "" || req.Price < 0 || req.StartDate == "" {
		s.handleError(w, r, apperrors.NewBadRequest("missing required fields for update", nil))
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

	userID := query.Get("user_id")
	if _, err := uuid.Parse(userID); err != nil {
		s.handleError(w, r, apperrors.NewBadRequest("invalid or missing user_id", err))
		return
	}

	periodStartStr := query.Get("period_start")
	periodStart, err := time.Parse("01-2006", periodStartStr)
	if err != nil {
		s.handleError(w, r, apperrors.NewBadRequest("invalid or missing period_start, use MM-YYYY format", err))
		return
	}

	periodEndStr := query.Get("period_end")
	periodEnd, err := time.Parse("01-2006", periodEndStr)
	if err != nil {
		s.handleError(w, r, apperrors.NewBadRequest("invalid or missing period_end, use MM-YYYY format", err))
		return
	}

	filter := dto.CostFilter{
		UserID:      userID,
		ServiceName: query.Get("service_name"),
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
