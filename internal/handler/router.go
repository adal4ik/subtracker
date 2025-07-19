package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
)

func Router(handlers Handlers) http.Handler {
	r := chi.NewRouter()

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(corsMiddleware.Handler)

	r.Post("/subscriptions", handlers.SubscriptionHandler.CreateSubscription)
	r.Get("/subscriptions", handlers.SubscriptionHandler.ListSubscriptions)
	r.Get("/subscriptions/{id}", handlers.SubscriptionHandler.GetSubscription)
	r.Put("/subscriptions/{id}", handlers.SubscriptionHandler.UpdateSubscription)
	r.Delete("/subscriptions/{id}", handlers.SubscriptionHandler.DeleteSubscription)
	r.Get("/subscriptions/cost", handlers.SubscriptionHandler.CalculateCost)

	r.Get("/swagger.json", handlers.SubscriptionHandler.ServeSwaggerJSON)

	return r
}
