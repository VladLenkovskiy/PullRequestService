package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"pr-service/internal/repository/pr"
	"pr-service/internal/repository/user_team"
	"syscall"
	"time"

	"pr-service/internal/config"
	"pr-service/internal/db"
	"pr-service/internal/logger"

	"pr-service/internal/service"

	prhand "pr-service/internal/handlers/pr_handlers"
	teamhand "pr-service/internal/handlers/team_handlers"
	userhand "pr-service/internal/handlers/user_handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	logger.Init(cfg.Logger.Level)

	log.Info().Msg("Starting PR Service")

	dbConfig := config.DatabaseConfig{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	database, err := db.NewPostgresDB(dbConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer database.Close()

	log.Info().Msg("Database connected successfully")

	teamRepo := user_team.NewUserTeamRepository(database)
	prRepo := pr.NewPRRepository(database)

	teamService := service.NewTeamService(teamRepo)
	userService := service.NewUserService(teamRepo, prRepo)
	prService := service.NewPRService(prRepo, teamRepo)

	teamHandler := teamhand.NewTeamHandler(teamService)
	userHandler := userhand.NewUserHandler(userService)
	prHandler := prhand.NewPRHandler(prService)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Timeout(30 * time.Second))

	teamHandler.RegisterRoutes(r)
	userHandler.RegisterRoutes(r)
	prHandler.RegisterRoutes(r)

	port := getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("port", port).Msg("Server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited gracefully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
