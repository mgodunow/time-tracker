package app

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"timeTracker/internal/config"
	"timeTracker/internal/controllers"
	"timeTracker/internal/repository"
	"timeTracker/internal/service"
)

type app struct {
	cfg     *config.Config
	handler *controllers.Handler
}

func NewApp() *app {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	config := config.MustLoad(dir)
	userService := service.NewUserService(repository.NewRepository(config.PostgresHost,
		config.AppPort,
		config.PostgresUser, config.PostgresPassword, config.PostgresDBName), config.GetByPassportDomain)
	handler := controllers.NewHandler(userService, slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	return &app{cfg: &config, handler: handler}
}

func (a *app) ListenAndServe() {
	log.Printf("Starting server on port %s", a.cfg.AppPort)
	if err := http.ListenAndServe(":"+a.cfg.AppPort, a.handler.Router()); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
