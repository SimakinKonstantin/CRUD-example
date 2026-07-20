package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"smartway/api"
	"strconv"
)

// Добавить код ошибки, сообщение об ошибке в HTTP ответ.
func addErrorInfo(w http.ResponseWriter, errorCode int, message string) error {
	w.WriteHeader(errorCode)
	respBody := api.ErrorResponse{Code: errorCode, Description: message}
	return json.NewEncoder(w).Encode(respBody)
}

// Проверяет корректность нового добавляемого пользователя.
func validateAddEmployee(e api.AddPatchEmployee) error {
	switch {
	case e.Name == "":
		return ValidationError{"имя не должно быть пустым"}
	case e.Surname == "":
		return ValidationError{"фамилия не должна быть пустой"}
	case e.Phone == "":
		return ValidationError{"телефон не должен быть пустым"}
	case e.CompanyId == 0:
		return ValidationError{"id компании должен быть целым положительным"}
	}

	if e.Passport.Type == "" || e.Passport.Number == "" {
		return ValidationError{"паспортные данные некорректны"}
	}
	if e.Department.Name == "" || e.Department.Phone == "" {
		return ValidationError{"данные отдела некорректны"}
	}
	return nil
}

// Проверяет, чтобы при обновлении через PATCH нельзя было обнулить поля пользователя.
func validatePatchEmployee(e api.Employee) error {
	switch {
	case e.Name == "":
		return ValidationError{"новое имя не должно быть пустым"}
	case e.Surname == "":
		return ValidationError{"новая фамилия не должна быть пустой"}
	case e.Phone == "":
		return ValidationError{"новый телефон не должен быть пустым"}
	case e.CompanyId == 0:
		return ValidationError{"новый id компании должен быть целым положительным"}
	}

	if e.Passport.Type == "" || e.Passport.Number == "" {
		return ValidationError{"новые паспортные должны быть корректны"}
	}
	if e.Department.Name == "" || e.Department.Phone == "" {
		return ValidationError{"новые данные отдела должны быть корректны"}
	}
	return nil
}

// AddEmployee godoc
// @Summary Добавить сотрудника
// @Description Создаёт сотрудника и возвращает его идентификатор.
// @Tags employees
// @Accept json
// @Produce json
// @Param employee body api.AddPatchEmployee true "Данные сотрудника"
// @Success 200 {object} api.AddEmployeeResponse
// @Failure 400 {object} api.ErrorResponse
// @Failure 422 {object} api.ErrorResponse
// @Router /employees [post]
func (deps *Dependencies) AddEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newEmployee api.AddPatchEmployee
	if err := json.NewDecoder(r.Body).Decode(&newEmployee); err != nil {
		errMsg := fmt.Sprintf("Ошибка разбора тела запроса: %s", err.Error())
		slog.Error(errMsg)

		if err = addErrorInfo(w, http.StatusBadRequest, errMsg); err != nil {
			slog.Error(fmt.Sprintf("Ошибка: %s; не удалось отправить JSON с информацией об ошибке: %s", err.Error(), errMsg))
		}
		return
	}

	if err := validateAddEmployee(newEmployee); err != nil {
		errMsg := fmt.Sprintf("Некорректное тело запроса: %s", err.Error())
		slog.Error(errMsg)

		if err = addErrorInfo(w, http.StatusBadRequest, errMsg); err != nil {
			slog.Error(fmt.Sprintf("Ошибка: %s; не удалось отправить JSON с информацией об ошибке: %s", err.Error(), errMsg))
		}
		return
	}

	employeeId, err := deps.Db.AddEmployee(r.Context(), newEmployee)
	if err != nil {
		errMsg := fmt.Sprintf("Ошибка добавления информации о сотруднике в БД: %s", err.Error())
		slog.Error(errMsg)

		if err = addErrorInfo(w, http.StatusUnprocessableEntity, errMsg); err != nil {
			slog.Error(fmt.Sprintf("Ошибка: %s; не удалось отправить JSON с информацией об ошибке: %s", err.Error(), errMsg))
		}
		return
	}

	if err = json.NewEncoder(w).Encode(api.AddEmployeeResponse{Id: employeeId}); err != nil {
		errMsg := fmt.Sprintf("Ошибка записи тела ответа: %s", err.Error())
		slog.Error(errMsg)

		if err = addErrorInfo(w, http.StatusUnprocessableEntity, errMsg); err != nil {
			slog.Error(fmt.Sprintf("Ошибка: %s; не удалось отправить JSON с информацией об ошибке: %s", err.Error(), errMsg))
		}
		return
	}
}

// DeleteEmployee godoc
// @Summary Удалить сотрудника
// @Description Удаляет сотрудника по идентификатору.
// @Tags employees
// @Produce json
// @Param id query int true "Идентификатор сотрудника"
// @Success 200 "Сотрудник удалён"
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Router /employees [delete]
func (deps *Dependencies) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	employeeIdStr := r.URL.Query().Get("id")
	employeeId, err := strconv.Atoi(employeeIdStr)
	if err != nil {
		errMsg := fmt.Sprintf("Ошибка: некорретный id сотрудника: %s", err.Error())
		slog.Error(errMsg)

		if err = addErrorInfo(w, http.StatusBadRequest, errMsg); err != nil {
			slog.Error(fmt.Sprintf("Ошибка: %s; не удалось отправить JSON с информацией об ошибке: %s", err.Error(), errMsg))
		}
		return
	}

	err = deps.Db.DeleteEmployee(r.Context(), employeeId)

	// Если при удалении информации о пользователе она не будет найдена.
	if errors.Is(err, sql.ErrNoRows) {
		errMsg := fmt.Sprintf("Сотрудник с id = %d не существует", employeeId)
		slog.Error(errMsg)

		if err = addErrorInfo(w, http.StatusNotFound, errMsg); err != nil {
			slog.Error(fmt.Sprintf("Ошибка: %s; не удалось отправить JSON с информацией об ошибке: %s", err.Error(), errMsg))
		}
	} else if err != nil {
		errMsg := fmt.Sprintf("Ошибка удаления информации о сотруднике из БД: %s", err.Error())
		slog.Error(errMsg)

		if err = addErrorInfo(w, http.StatusBadRequest, errMsg); err != nil {
			slog.Error(fmt.Sprintf("Ошибка: %s; не удалось отправить JSON с информацией об ошибке: %s", err.Error(), errMsg))
		}
		return
	}
}

// GetEmployeesByCompanyId godoc
// @Summary Получить сотрудников компании
// @Description Возвращает сотрудников по идентификатору компании.
// @Tags employees
// @Produce json
// @Param id query int true "Идентификатор компании"
// @Success 200 {array} api.Employee
// @Success 204 "Сотрудники не найдены"
// @Failure 400 {object} api.ErrorResponse
// @Failure 422 {object} api.ErrorResponse
// @Router /company/employees [get]
func (deps *Dependencies) GetEmployeesByCompanyId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	companyIdStr := r.URL.Query().Get("id")
	companyId, err := strconv.Atoi(companyIdStr)
	if err != nil {
		errMsg := fmt.Sprintf("Ошибка: некорретный id компании: %s", err.Error())
		slog.Error(errMsg)

		if err = addErrorInfo(w, http.StatusBadRequest, errMsg); err != nil {
			slog.Error(fmt.Sprintf("Ошибка: %s; не удалось отправить JSON с информацией об ошибке: %s", err.Error(), errMsg))
		}
		return
	}

	employees, err := deps.Db.GetEmployeesByCompany(r.Context(), companyId)
	if err != nil {
		errMsg := fmt.Sprintf("Ошибка получения информации о сотрудниках по id компании из БД: %s", err.Error())
		slog.Error(errMsg)

		if err = addErrorInfo(w, http.StatusUnprocessableEntity, errMsg); err != nil {
			slog.Error(fmt.Sprintf("Ошибка: %s; не удалось отправить JSON с информацией об ошибке: %s", err.Error(), errMsg))
		}
		return
	}

	if len(employees) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err = json.NewEncoder(w).Encode(employees); err != nil {
		errMsg := fmt.Sprintf("Ошибка записи тела ответа: %s", err.Error())
		slog.Error(errMsg)

		if err = addErrorInfo(w, http.StatusUnprocessableEntity, errMsg); err != nil {
			slog.Error(fmt.Sprintf("Ошибка: %s; не удалось отправить JSON с информацией об ошибке: %s", err.Error(), errMsg))
		}
		return
	}
}

// GetEmployeesByDepartmentName godoc
// @Summary Получить сотрудников отдела
// @Description Возвращает сотрудников по названию отдела.
// @Tags employees
// @Produce json
// @Param name query string true "Название отдела"
// @Success 200 {array} api.Employee
// @Success 204 "Сотрудники не найдены"
// @Failure 400 {object} api.ErrorResponse
// @Failure 422 {object} api.ErrorResponse
// @Router /department/employees [get]
func (deps *Dependencies) GetEmployeesByDepartmentName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	departmentName := r.URL.Query().Get("name")
	if departmentName == "" {
		errMsg := "Ошибка: имя отдела не может быть пустым"
		slog.Error(errMsg)

		if err := addErrorInfo(w, http.StatusBadRequest, errMsg); err != nil {
			slog.Error(fmt.Sprintf("Ошибка: %s; не удалось отправить JSON с информацией об ошибке: %s", err.Error(), errMsg))
		}
		return
	}

	employees, err := deps.Db.GetEmployeesByDepartment(r.Context(), departmentName)
	if err != nil {
		errMsg := fmt.Sprintf("Ошибка получения информации о сотрудниках по id отдела из БД: %s", err.Error())
		slog.Error(errMsg)

		if err = addErrorInfo(w, http.StatusUnprocessableEntity, errMsg); err != nil {
			slog.Error(fmt.Sprintf("Ошибка: %s; не удалось отправить JSON с информацией об ошибке: %s", err.Error(), errMsg))
		}
		return
	}

	if len(employees) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err = json.NewEncoder(w).Encode(employees); err != nil {
		errMsg := fmt.Sprintf("Ошибка записи тела ответа: %s", err.Error())
		slog.Error(errMsg)

		if err = addErrorInfo(w, http.StatusUnprocessableEntity, errMsg); err != nil {
			slog.Error(fmt.Sprintf("Ошибка: %s; не удалось отправить JSON с информацией об ошибке: %s", err.Error(), errMsg))
		}
		return
	}
}

// PatchEmployee godoc
// @Summary Обновить сотрудника
// @Description Частично обновляет сотрудника по идентификатору.
// @Tags employees
// @Accept json
// @Produce json
// @Param id query int true "Идентификатор сотрудника"
// @Param employee body api.AddPatchEmployee true "Изменяемые поля сотрудника"
// @Success 200 "Сотрудник обновлён"
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 422 {object} api.ErrorResponse
// @Router /employees [patch]
func (deps *Dependencies) PatchEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	employeeIdStr := r.URL.Query().Get("id")
	employeeId, err := strconv.Atoi(employeeIdStr)
	if err != nil {
		errMsg := fmt.Sprintf("Ошибка: некорретный id сотрудника: %s", err.Error())
		slog.Error(errMsg)

		if err = addErrorInfo(w, http.StatusBadRequest, errMsg); err != nil {
			slog.Error(fmt.Sprintf("Ошибка: %s; не удалось отправить JSON с информацией об ошибке: %s", err.Error(), errMsg))
		}
		return
	}

	employee, err := deps.Db.GetEmployeeById(r.Context(), employeeId)

	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		errMsg := fmt.Sprintf("Ошибка получения информации о сотруднике из БД: %s", err.Error())
		slog.Error(errMsg)

		if err = addErrorInfo(w, http.StatusUnprocessableEntity, errMsg); err != nil {
			slog.Error(fmt.Sprintf("Ошибка: %s; не удалось отправить JSON с информацией об ошибке: %s", err.Error(), errMsg))
		}
		return
	}

	// Перезаписываем текущие значения значениями, которые были даны в JSON запроса.
	err = json.NewDecoder(r.Body).Decode(&employee)
	if err != nil {
		errMsg := fmt.Sprintf("Ошибка разбора тела запроса: %s", err.Error())
		slog.Error(errMsg)

		if err = addErrorInfo(w, http.StatusBadRequest, errMsg); err != nil {
			slog.Error(fmt.Sprintf("Ошибка: %s; не удалось отправить JSON с информацией об ошибке: %s", err.Error(), errMsg))
		}
		return
	}

	if err := validatePatchEmployee(employee); err != nil {
		errMsg := fmt.Sprintf("Некорректное тело запроса: %s", err.Error())
		slog.Error(errMsg)

		if err = addErrorInfo(w, http.StatusBadRequest, errMsg); err != nil {
			slog.Error(fmt.Sprintf("Ошибка: %s; не удалось отправить JSON с информацией об ошибке: %s", err.Error(), errMsg))
		}
		return
	}

	// Обновляем значения в БД.
	err = deps.Db.UpdateEmployee(r.Context(), employeeId, employee)
	if err != nil {
		errMsg := fmt.Sprintf("Ошибка обновления информации о сотруднике в БД: %s", err.Error())
		slog.Error(errMsg)

		if err = addErrorInfo(w, http.StatusUnprocessableEntity, errMsg); err != nil {
			slog.Error(fmt.Sprintf("Ошибка: %s; не удалось отправить JSON с информацией об ошибке: %s", err.Error(), errMsg))
		}
		return
	}
}
