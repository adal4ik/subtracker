package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router(handlers Handlers) http.Handler {
	r := chi.NewRouter()
	r.Post("/subscription", handlers.SubscriptionHandler.CreateSubscription)
	r.Get("/subscriptions", handlers.SubscriptionHandler.ListSubscriptions)
	// r.Get("/subscription/{id}", handlers.SubscriptionHandler.GetSubscription)
	// r.Put("/subscription/{id}", handlers.SubscriptionHandler.UpdateSubscription)
	// r.Delete("/subscription/{id}", handlers.SubscriptionHandler.DeleteSubscription)

	return r
}
