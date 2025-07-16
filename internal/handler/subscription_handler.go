package handler

import (
	"subtracker/internal/service"
	"subtracker/pkg/logger"
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
