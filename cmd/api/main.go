package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aggregator_db/internal/config"
	httpHandler "aggregator_db/internal/handler/http"
	"aggregator_db/internal/repository/postgres"
	"aggregator_db/internal/service"
	"aggregator_db/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "aggregator_db/docs"
)

// @title           Subscription Service API
// @version         1.0
// @description     REST API для управления онлайн-подписками пользователей
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @schemes http https
func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация логгера
	appLogger := logger.New(cfg.LogLevel)
	appLogger.Info("Starting subscription service",
		"port", cfg.ServerPort,
	)

	// Подключение к БД
	dbPool, err := pgxpool.New(context.Background(), cfg.DSN())
	if err != nil {
		appLogger.Error("Failed to connect to database", "error", err.Error())
		os.Exit(1)
	}
	defer dbPool.Close()

	// Проверка подключения
	if err := dbPool.Ping(context.Background()); err != nil {
		appLogger.Error("Failed to ping database", "error", err.Error())
		os.Exit(1)
	}
	appLogger.Info("Successfully connected to database")

	// Инициализация слоев приложения
	subscriptionRepo := postgres.NewSubscriptionRepository(dbPool)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo, appLogger)

	// Настройка роутера
	router := httpHandler.SetupRouter(subscriptionService, appLogger)

	// Graceful shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: router,
	}

	go func() {
		appLogger.Info("Server is running", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error("Failed to start server", "error", err.Error())
			os.Exit(1)
		}
	}()

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Error("Server forced to shutdown", "error", err.Error())
	}

	appLogger.Info("Server exited")
}
