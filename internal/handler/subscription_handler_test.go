package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"subtracker/internal/domain"
	"subtracker/internal/domain/dto"
	"subtracker/internal/service/mocks"
	"subtracker/pkg/apperrors"
	"subtracker/pkg/logger"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateSubscription(t *testing.T) {
	mockService := new(mocks.SubscriptionServiceInterface)
	handler := NewSubscriptionHandler(mockService, logger.NewNopLogger())

	t.Run("Success", func(t *testing.T) {
		reqBody := dto.CreateSubscriptionRequest{
			ServiceName: "Netflix",
			Price:       500,
			UserID:      uuid.New().String(),
			StartDate:   "01-2025",
		}
		body, _ := json.Marshal(reqBody)

		mockService.On("CreateSubscription", mock.Anything, mock.AnythingOfType("domain.Subscription")).Return(nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		handler.CreateSubscription(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Validation Error", func(t *testing.T) {
		reqBody := dto.CreateSubscriptionRequest{Price: -100}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		handler.CreateSubscription(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockService.AssertNotCalled(t, "CreateSubscription")
	})
}

func TestListSubscriptions(t *testing.T) {
	mockService := new(mocks.SubscriptionServiceInterface)
	handler := NewSubscriptionHandler(mockService, logger.NewNopLogger())

	t.Run("Success", func(t *testing.T) {
		mockResponse := []domain.Subscription{{ID: uuid.New()}}
		mockService.On("ListSubscriptions", mock.Anything, mock.AnythingOfType("dto.SubscriptionFilter")).Return(mockResponse, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/subscriptions?limit=5", nil)
		rr := httptest.NewRecorder()
		handler.ListSubscriptions(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var responseBody []dto.SubscriptionResponse
		json.Unmarshal(rr.Body.Bytes(), &responseBody)
		assert.Len(t, responseBody, 1)
		mockService.AssertExpectations(t)
	})

	t.Run("Validation Error on Filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/subscriptions?limit=200", nil)
		rr := httptest.NewRecorder()
		handler.ListSubscriptions(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockService.AssertNotCalled(t, "ListSubscriptions")
	})
}

func TestGetSubscription(t *testing.T) {
	mockService := new(mocks.SubscriptionServiceInterface)
	handler := NewSubscriptionHandler(mockService, logger.NewNopLogger())
	router := chi.NewRouter()
	router.Get("/subscriptions/{id}", handler.GetSubscription)

	t.Run("Success", func(t *testing.T) {
		testID := uuid.New()
		mockResponse := domain.Subscription{ID: testID}
		mockService.On("GetSubscription", mock.Anything, testID.String()).Return(mockResponse, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/subscriptions/"+testID.String(), nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var respBody dto.SubscriptionResponse
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Equal(t, testID.String(), respBody.ID)
		mockService.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		testID := uuid.New().String()
		repoErr := apperrors.NewNotFound("not found", nil)
		mockService.On("GetSubscription", mock.Anything, testID).Return(domain.Subscription{}, repoErr).Once()

		req := httptest.NewRequest(http.MethodGet, "/subscriptions/"+testID, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid ID Format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/subscriptions/not-a-uuid", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockService.AssertNotCalled(t, "GetSubscription")
	})
}

func TestUpdateSubscription(t *testing.T) {
	mockService := new(mocks.SubscriptionServiceInterface)
	handler := NewSubscriptionHandler(mockService, logger.NewNopLogger())
	router := chi.NewRouter()
	router.Put("/subscriptions/{id}", handler.UpdateSubscription)

	t.Run("Success", func(t *testing.T) {
		testID := uuid.New()
		reqBody := dto.UpdateSubscriptionRequest{ServiceName: "New Name", Price: 123, StartDate: "02-2025"}
		body, _ := json.Marshal(reqBody)

		mockService.On("UpdateSubscription", mock.Anything, mock.AnythingOfType("domain.Subscription")).Return(nil).Once()

		req := httptest.NewRequest(http.MethodPut, "/subscriptions/"+testID.String(), bytes.NewReader(body))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Validation Error", func(t *testing.T) {
		testID := uuid.New().String()
		reqBody := dto.UpdateSubscriptionRequest{Price: -1}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/subscriptions/"+testID, bytes.NewReader(body))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockService.AssertNotCalled(t, "UpdateSubscription")
	})
}

func TestDeleteSubscription(t *testing.T) {
	mockService := new(mocks.SubscriptionServiceInterface)
	handler := NewSubscriptionHandler(mockService, logger.NewNopLogger())
	router := chi.NewRouter()
	router.Delete("/subscriptions/{id}", handler.DeleteSubscription)

	t.Run("Success", func(t *testing.T) {
		testID := uuid.New().String()
		mockService.On("DeleteSubscription", mock.Anything, testID).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/subscriptions/"+testID, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		testID := uuid.New().String()
		repoErr := apperrors.NewNotFound("not found", nil)
		mockService.On("DeleteSubscription", mock.Anything, testID).Return(repoErr).Once()

		req := httptest.NewRequest(http.MethodDelete, "/subscriptions/"+testID, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockService.AssertExpectations(t)
	})
}

func TestCalculateCost(t *testing.T) {
	mockService := new(mocks.SubscriptionServiceInterface)
	handler := NewSubscriptionHandler(mockService, logger.NewNopLogger())

	t.Run("Success", func(t *testing.T) {
		mockService.On("CalculateCost", mock.Anything, mock.AnythingOfType("dto.CostFilter")).Return(1500, nil).Once()

		url := "/subscriptions/cost?user_id=" + uuid.New().String() + "&period_start=01-2025&period_end=03-2025"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()
		handler.CalculateCost(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var respBody dto.CostResponse
		json.Unmarshal(rr.Body.Bytes(), &respBody)
		assert.Equal(t, 1500, respBody.TotalCost)
		mockService.AssertExpectations(t)
	})

	t.Run("Validation Error", func(t *testing.T) {
		url := "/subscriptions/cost?user_id=not-a-uuid&period_start=01-2025&period_end=03-2025"
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()
		handler.CalculateCost(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockService.AssertNotCalled(t, "CalculateCost")
	})
}
