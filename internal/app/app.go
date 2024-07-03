package app

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"timeTracker/internal/config"
	"timeTracker/internal/controllers"
	"timeTracker/internal/repository"
	"timeTracker/internal/service"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

// TODO: add path to migrations to config
func (a *app) Migrate() {
	path, err := filepath.Abs("./migrations")
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}
	sourceURL := fmt.Sprintf("file://%s", filepath.ToSlash(path))
	log.Printf("Using migrations from: %s", sourceURL)

	m, err := migrate.New(sourceURL, fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		a.cfg.PostgresUser, a.cfg.PostgresPassword, a.cfg.PostgresHost,
		a.cfg.PostgresPort, a.cfg.PostgresDBName))
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to apply migrations: %v", err)
	}
	log.Println("Migrations applied successfully")
}
func (a *app) ListenAndServe() {
	log.Printf("Starting server on port %s", a.cfg.AppPort)
	if err := http.ListenAndServe(":"+a.cfg.AppPort, a.handler.Router()); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
