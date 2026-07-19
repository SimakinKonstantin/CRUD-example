package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	_ "smartway/docs"
	"smartway/internal/config"
	"smartway/internal/db"
	"smartway/internal/handlers"
	"smartway/internal/server"
	"syscall"
	"time"
)

// @title Employee Service API
// @version 1.0
// @description REST API для управления сотрудниками компаний.
// @host localhost:8080
// @BasePath /
// @schemes http
func main() {

	// Подключаемся к БД.
	dbProcessor, err := db.CreatePgxProcessor(os.Getenv("POSTGRES_CONNECT_URL"))
	if err != nil {
		slog.Error(fmt.Sprintf("Не удалось подключиться к БД: %s", err.Error()))
		return
	}

	// Инициализируем все необходимые зависимости.
	appDependencies := handlers.CreateDependencies(dbProcessor)
	defer appDependencies.Close()

	// Проверяем зависимости.
	err = appDependencies.Check()
	if err != nil {
		slog.Error(fmt.Sprintf("Ошибка при проверке зависимостей: %s", err.Error()))
		return
	}

	httpServer := server.Create(config.ServerAddr, appDependencies)
	serverErr := make(chan error, 1)
	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		serverErr <- httpServer.ListenAndServe()
	}()

	select {
	case err = <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Ошибка HTTP-сервера", "error", err)
		}
		return
	case <-shutdownSignal.Done():
		slog.Info("Cервер останавливается")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("Не удалось корректно остановить HTTP-сервер", "error", err)
		if closeErr := httpServer.Close(); closeErr != nil {
			slog.Error("Не удалось принудительно остановить HTTP-сервер", "error", closeErr)
		}
		return
	}

	if err = <-serverErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Ошибка при завершении HTTP-сервера", "error", err)
		return
	}

	slog.Info("Сервер остановлен")
}
