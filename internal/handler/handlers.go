package handler

import (
	"subtracker/internal/service"
	"subtracker/pkg/logger"
)

type Handlers struct {
	SubscriptionHandler *SubscriptionHandler
}

func NewHandlers(service *service.Service, logger logger.Logger) *Handlers {
	return &Handlers{
		SubscriptionHandler: NewSubscriptionHandler(service.SubscriptionService, logger),
	}
}
