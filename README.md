# Web-сервис сотрудников, сделанный на Golang

## Docker Compose

Для запуска API и PostgreSQL одной командой выполните:

```bash
docker compose up --build
```

После запуска API доступно по адресу `http://localhost:8080`, PostgreSQL — на внутреннем адресе `db:5432` сети Compose. Перед запуском API контейнер выполняет `goose up` для всех миграций из `migrations`.

Интерактивная документация Swagger доступна по адресу `http://localhost:8080/swagger/`.

Состояние API и соединения с PostgreSQL можно проверить запросом `GET http://localhost:8080/health`.

Чтобы остановить сервисы, выполните `docker compose down`. Для полного сброса данных базы используйте `docker compose down -v`.

## Функционал
1. Возможность добавления сотрудников, в ответ должен приходить Id добавленного сотрудника;
2. Возможность удаления сотрудников по Id;
3. Возможность выводить список сотрудников для указанной компании;
4. Возможность все список сотрудников для указанного отдела компании;
5. Возможность изменения определенных полей сотрудников по его Id.

Модель сотрудника:
   ```
   {
       Id int
       Name string
       Surname string
       Phone string
       CompanyId int
       Passport {
           Type string
           Number string
       }
       Department {
           Name string
           Phone string
       }
   }
   ```
   
### Инструкция по запуску:
1. Выполнить: `docker compose up`;
2. Перейти на `http://localhost:8080/swagger/index.html`.
