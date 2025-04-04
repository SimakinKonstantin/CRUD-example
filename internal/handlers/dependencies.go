package handlers

import "smartway/internal/db"

// Структура с зависимостями, доступ к которым есть у хендлеров.
type Dependencies struct {
	Db db.DbProcessor
}

// Создание структуры.
func CreateDependencies(Db db.DbProcessor) *Dependencies {
	return &Dependencies{Db: Db}
}

// Проверка, что все зависимости корректно подключены.
func (deps *Dependencies) Check() error {
	return deps.Db.Ping()
}

// Закрытие всех зависимостей.
func (deps *Dependencies) Close() {
	deps.Db.CloseConnection()
}
