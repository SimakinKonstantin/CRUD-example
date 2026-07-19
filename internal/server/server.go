package server

import (
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"net/http"
	"smartway/internal/handlers"
	"time"
)

// Создание сервера;
// addr - адрес сервера;
// deps - зависимости, необбходимые хендлерам для работы.
func Create(addr string, deps *handlers.Dependencies) *http.Server {
	router := mux.NewRouter()

	router.HandleFunc("/employees", deps.AddEmployee).Methods("POST")
	router.HandleFunc("/employees", deps.DeleteEmployee).Methods("DELETE")
	router.HandleFunc("/employees", deps.PatchEmployee).Methods("PATCH")
	router.HandleFunc("/company/employees", deps.GetEmployeesByCompanyId).Methods("GET")
	router.HandleFunc("/department/employees", deps.GetEmployeesByDepartmentName).Methods("GET")
	router.HandleFunc("/health", deps.Health).Methods("GET")
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	return &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
