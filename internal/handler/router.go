package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router(handlers Handlers) http.Handler {
	r := chi.NewRouter()

	return r
}
