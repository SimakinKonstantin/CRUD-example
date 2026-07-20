package db

import (
	"context"
	"errors"
	"fmt"
	"smartway/api"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Интерфейс для работы с БД.
type DbProcessor interface {
	CloseConnection()
	Ping() error
	AddEmployee(ctx context.Context, employee api.AddPatchEmployee) (employeeId int, err error)
	DeleteEmployee(ctx context.Context, employeeId int) error
	GetEmployeesByCompany(ctx context.Context, companyId int) ([]api.Employee, error)
	GetEmployeesByDepartment(ctx context.Context, departmentName string) ([]api.Employee, error)
	GetEmployeeById(ctx context.Context, employeeId int) (api.Employee, error)
	UpdateEmployee(ctx context.Context, employeeId int, newValue api.Employee) error
}

// Реализация DbProcessor, которая работает с pgx.
type PgxProcessor struct {
	connection *pgxpool.Pool
}

// Создание соединение.
func CreatePgxProcessor(dbUrl string) (*PgxProcessor, error) {
	connection, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		return nil, err
	}

	return &PgxProcessor{connection: connection}, nil
}

// Закрытие соединения.
func (db *PgxProcessor) CloseConnection() {
	db.connection.Close()
}

// Пинг БД.
func (db *PgxProcessor) Ping() error {
	return db.connection.Ping(context.Background())
}

// Добавление нового сотрудника в БД.
func (db *PgxProcessor) AddEmployee(ctx context.Context, employee api.AddPatchEmployee) (employeeId int, err error) {
	// Добавление взаимодействует с 3 таблицами: Passports, Departments, Employees, поэтому используется транзакция.
	tx, err := db.connection.Begin(ctx)
	if err != nil {
		return -1, err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback(ctx))
		}
	}()

	// Для каждого нового сотрудника добавляется информация о паспорте в таблицу с паспортами.
	var passportId int
	err = tx.QueryRow(
		ctx,
		"INSERT INTO Passports (type, number) VALUES ($1, $2) RETURNING id;",
		employee.Passport.Type, employee.Passport.Number,
	).Scan(&passportId)
	if err != nil {
		return -1, err
	}

	// Для нового сотрудника нужно сначала проверить, есть ли отдел с таким названием, если нет - добавить новый отдел.
	var departmentId int
	var departmentPhone string
	err = tx.QueryRow(
		ctx,
		"SELECT id, phone FROM Departments WHERE name = $1;", employee.Department.Name,
	).Scan(&departmentId, &departmentPhone)

	// Если нужно добавить сотрудника в существующий отдел, но указан неверный номер отдела, возвращаем ошибку.
	if err == nil && employee.Department.Phone != departmentPhone {
		errMsg := fmt.Sprintf("Отдел %s имеет иной телефон - %s", employee.Department.Name, departmentPhone)
		return -1, errors.New(errMsg)
	}

	// Добавляем отдел при его отсутствии, в итоге в departmentId в любом случае будет лежать корректный id отдела.
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(
			ctx,
			"INSERT INTO Departments (name, phone) VALUES ($1, $2) RETURNING id;",
			employee.Department.Name, employee.Department.Phone,
		).Scan(&departmentId)
		if err != nil {
			return -1, err
		}
	}

	employeeAddQuery :=
		`INSERT INTO Employees (name, surname, phone, company_id, passport_id, department_id)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;`

	// Добавляем сотрудника в БД, возвращаем сгенерированный Id.
	err = tx.QueryRow(
		ctx,
		employeeAddQuery,
		employee.Name, employee.Surname, employee.Phone, employee.CompanyId, passportId, departmentId).Scan(&employeeId)
	if err != nil {
		return -1, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return -1, err
	}

	return employeeId, nil
}

// Удаление сотрудника по id из БД.
func (db *PgxProcessor) DeleteEmployee(ctx context.Context, employeeId int) (err error) {
	// Удаление взаимодействует с 3 таблицами: Passports, Employees, поэтому используется транзакция.
	// В задании указано, что есть Departments, в которых может быть несколько сотрудников. Про удаление отделов не сказано,
	// поэтому при удалении пользователей отделы не затрагиваются.

	tx, err := db.connection.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback(ctx))
		}
	}()

	var passportId int

	// Получаем id FK для удаления из таблицы Passports.
	err = tx.QueryRow(
		ctx,
		"SELECT passport_id FROM Employees WHERE id=$1;",
		employeeId,
	).Scan(&passportId)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		"DELETE FROM Employees WHERE id=$1;", employeeId,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		"DELETE FROM Passports WHERE id=$1;", passportId,
	)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Получение сотрудников по id компании.
func (db *PgxProcessor) GetEmployeesByCompany(ctx context.Context, companyId int) ([]api.Employee, error) {
	dbQuery :=
		`SELECT e.id id, e.name name, e.surname surname, e.phone phone, e.company_id company_id,
p.type passport_type, p.number passport_number,  d.name department_name, d.phone department_phone FROM Employees AS e
JOIN Passports AS p
ON e.passport_id = p.id
JOIN Departments AS d
ON e.department_id = d.id
WHERE e.company_id=$1;`

	rows, err := db.connection.Query(ctx, dbQuery, companyId)
	if err != nil {
		return nil, err
	}

	result := make([]api.Employee, 0)

	for rows.Next() {
		var employee api.Employee
		err = rows.Scan(&employee.Id, &employee.Name, &employee.Surname, &employee.Phone, &employee.CompanyId,
			&employee.Passport.Type, &employee.Passport.Number, &employee.Department.Name, &employee.Department.Phone)
		if err != nil {
			return nil, err
		}

		result = append(result, employee)
	}

	return result, nil
}

// Получение сотрудников по названию отдела.
func (db *PgxProcessor) GetEmployeesByDepartment(ctx context.Context, departmentName string) ([]api.Employee, error) {
	dbQuery :=
		`SELECT e.id id, e.name name, e.surname surname, e.phone phone, e.company_id company_id,
p.type passport_type, p.number passport_number,  d.name department_name, d.phone department_phone FROM Employees AS e
JOIN Passports AS p
ON e.passport_id = p.id
JOIN Departments AS d
ON e.department_id = d.id
WHERE d.name=$1;`

	rows, err := db.connection.Query(
		ctx,
		dbQuery,
		departmentName,
	)
	if err != nil {
		return nil, err
	}

	result := make([]api.Employee, 0)

	for rows.Next() {
		var employee api.Employee
		err = rows.Scan(&employee.Id, &employee.Name, &employee.Surname, &employee.Phone, &employee.CompanyId,
			&employee.Passport.Type, &employee.Passport.Number, &employee.Department.Name, &employee.Department.Phone)
		if err != nil {
			return nil, err
		}

		result = append(result, employee)
	}

	return result, nil
}

// Получение сотрудника по id.
func (db *PgxProcessor) GetEmployeeById(ctx context.Context, employeeId int) (api.Employee, error) {
	dbQuery :=
		`SELECT e.id id, e.name name, e.surname surname, e.phone phone, e.company_id company_id,
p.type passport_type, p.number passport_number,  d.name department_name, d.phone department_phone FROM Employees AS e
JOIN Passports AS p
ON e.passport_id = p.id
JOIN Departments AS d
ON e.department_id = d.id
WHERE e.id=$1;`

	var employee api.Employee

	err := db.connection.QueryRow(
		ctx,
		dbQuery,
		employeeId,
	).Scan(&employee.Id, &employee.Name, &employee.Surname, &employee.Phone, &employee.CompanyId,
		&employee.Passport.Type, &employee.Passport.Number, &employee.Department.Name, &employee.Department.Phone)
	if err != nil {
		return employee, err
	}

	return employee, nil
}

// Обновление информации о сотруднике;
// employeeId - id сотрудника;
// newValue - новое значение информации о сотруднике.
func (db *PgxProcessor) UpdateEmployee(ctx context.Context, employeeId int, newValue api.Employee) (err error) {
	// Обновление может затронуть несколько таблиц, поэтому используется транзакция.
	tx, err := db.connection.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback(ctx))
		}
	}()

	// Получаем значение FK сотрудника.
	var passportId, departmentId int
	err = tx.QueryRow(
		ctx,
		"SELECT passport_id, department_id FROM Employees WHERE id=$1;",
		employeeId,
	).Scan(&passportId, &departmentId)

	if err != nil {
		return err
	}

	// Обновляем паспортные данные.
	_, err = tx.Exec(
		ctx,
		"UPDATE Passports SET type = $1, number = $2 WHERE id=$3;",
		newValue.Passport.Type, newValue.Passport.Number, passportId)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		"UPDATE Departments SET name = $1, phone = $2 WHERE id=$3;",
		newValue.Department.Name, newValue.Department.Phone, departmentId)
	if err != nil {
		return err
	}

	// Обновляем информацию о пользователе.
	_, err = tx.Exec(
		ctx,
		"UPDATE Employees SET name = $1, surname = $2, phone = $3, company_id = $4 WHERE id=$5;",
		newValue.Name, newValue.Surname, newValue.Phone, newValue.CompanyId, employeeId)

	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
