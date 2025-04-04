package api

// Модель сотрудника.
type Employee struct {
	Id        int    `json:"Id"`
	Name      string `json:"Name"`
	Surname   string `json:"Surname"`
	Phone     string `json:"Phone"`
	CompanyId int    `json:"CompanyId"`
	Passport  struct {
		Type   string `json:"Type"`
		Number string `json:"Number"`
	} `json:"Passport"`
	Department struct {
		Name  string `json:"Name"`
		Phone string `json:"Phone"`
	} `json:"Department"`
}

// Модель запроса на добавление/обновление сотрудника. В отличие от Employee не имеет поля Id.
// Поля Id для всех сущностей сервиса задаются автоматически через SERIAL.
type AddPatchEmployee struct {
	Name      string `json:"Name"`
	Surname   string `json:"Surname"`
	Phone     string `json:"Phone"`
	CompanyId int    `json:"CompanyId"`
	Passport  struct {
		Type   string `json:"Type"`
		Number string `json:"Number"`
	} `json:"Passport"`
	Department struct {
		Name  string `json:"Name"`
		Phone string `json:"Phone"`
	} `json:"Department"`
}

// Модель ответа на добавление сотрудника.
type AddEmployeeResponse struct {
	Id int `json:"Id"`
}

// Модель сообщения об ошибке.
type ErrorResponse struct {
	Code        int    `json:"Code"`
	Description string `json:"Description"`
}
