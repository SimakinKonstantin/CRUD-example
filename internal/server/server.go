package server

import (
	"github.com/gorilla/mux"
	"net/http"
	"smartway/internal/handlers"
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

	return &http.Server{
		Addr:    addr,
		Handler: router,
	}
}
