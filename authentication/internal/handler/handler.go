package handler

import "github.com/go-chi/chi/v5"

type AppHandler interface {
	RegisterRoutes(r chi.Router) chi.Router
}

func RegisterRoutes(router chi.Router, handlers ...AppHandler) {
	for _, handler := range handlers {
		handler.RegisterRoutes(router)
	}
}
