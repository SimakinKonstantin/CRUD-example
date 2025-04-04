-- +goose Up
CREATE TABLE Passports (
                           id SERIAL PRIMARY KEY,
                           type TEXT NOT NULL,
                           number TEXT UNIQUE NOT NULL
);

CREATE TABLE Departments (
                             id SERIAL PRIMARY KEY,
                             name TEXT NOT NULL UNIQUE,
                             phone TEXT NOT NULL UNIQUE
);

CREATE TABLE Employees (
                           id SERIAL PRIMARY KEY,
                           name TEXT NOT NULL,
                           surname TEXT NOT NULL,
                           phone TEXT NOT NULL UNIQUE ,
                           company_id INT NOT NULL,
                           passport_id INT NOT NULL UNIQUE REFERENCES Passports(id) ON DELETE CASCADE ON UPDATE CASCADE,
                           department_id INT NOT NULL REFERENCES Departments(id) ON UPDATE CASCADE
);

-- +goose Down
DROP TABLE Employees;
DROP TABLE Departments;
DROP TABLE Passports;