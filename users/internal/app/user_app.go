package app

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nimbo1999/financeiro/users/internal/handler"
	"github.com/nimbo1999/financeiro/users/internal/repositories"
	"github.com/nimbo1999/financeiro/users/internal/services"
	"gorm.io/gorm"
)

type App struct {
	db *gorm.DB
}

func New(db *gorm.DB) *App {
	return &App{db: db}
}

func (a *App) Run(port string) error {
	if port == "" {
		port = "8080"
	}

	userRepository := repositories.NewUserRepository(a.db)
	userService := services.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)

	mux := chi.NewMux()
	mux.Route("/", userHandler.RegisterRoutes)

	fmt.Println("Starting server on port:", port)
	return http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
}
