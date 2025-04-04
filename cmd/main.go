package main

import (
	"fmt"
	"log/slog"
	"os"
	"smartway/internal/config"
	"smartway/internal/db"
	"smartway/internal/handlers"
	"smartway/internal/server"
)

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

	server := server.Create(config.ServerAddr, appDependencies)
	server.ListenAndServe()
}
