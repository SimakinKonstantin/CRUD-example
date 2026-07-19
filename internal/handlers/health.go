package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"smartway/api"
)

// Health godoc
// @Summary Проверить состояние сервиса
// @Description Проверяет доступность API и соединение с PostgreSQL.
// @Tags health
// @Produce json
// @Success 200 {object} api.HealthResponse
// @Failure 503 {object} api.ErrorResponse
// @Router /health [get]
func (deps *Dependencies) Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := deps.Check(); err != nil {
		slog.Error("Проверка состояния PostgreSQL завершилась ошибкой", "error", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		if encodeErr := json.NewEncoder(w).Encode(api.ErrorResponse{
			Code:        http.StatusServiceUnavailable,
			Description: "PostgreSQL недоступен",
		}); encodeErr != nil {
			slog.Error("Не удалось записать ответ healthcheck", "error", encodeErr)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(api.HealthResponse{Status: "ok"}); err != nil {
		slog.Error("Не удалось записать ответ healthcheck", "error", err)
	}
}
