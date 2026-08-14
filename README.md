##  Быстрый запуск (Docker Compose)

Для запуска вам потребуется только установленный Docker.

**1. Клонируйте репозиторий:**
```bash
git clone https://github.com/your-username/NTEC_task_RESTAPI.git cd NTEC_task_RESTAPI
```
## Настройте окружение (опционально):
 2. Проект может использовать переменные из файла .env (если он есть в корне проекта). По умолчанию заданы следующие fallback-значения:
```bash
DATABASE_URL = postgres://admin:adminpassword@db:5432/task_db?sslmode=disable

JWT_SECRET = super_secret_key

SERVER_PORT = :8080

WORKER_INTERVAL = 1m
```
3. Запустите инфраструктуру:

```bash
docker compose up --build
```

🧪 Тестирование
Для запуска тестов локально:
```bash
go test -v ./internal/service/...
```

## 📖 Документация API

### Аутентификация
1. Регистрация
```bash
POST /register

Body: {"username": "user1", "password": "password123"}

Response (201): {"id": 1}
```
2. Логин
```bash
POST /login

Body: {"username": "user1", "password": "password123"}

Response (200): {"token": "eyJhbGciOiJIUzI1NiIs..."}
```

## Задачи (Требуется JWT Токен)

Для доступа передайте заголовок: Authorization: Bearer <token>

1. Создать задачу
```bash
POST /tasks

Body:

JSON
{
  "title": "Настроить CI/CD",
  "description": "Написать GitHub Actions workflow",
  "deadline": "2026-12-31T23:59:59Z",
  "responsible_id": 2
}
(Если responsible_id не указан, ответственным автоматически назначается создатель)
```
Response (201): {"id": 1}

2. Получить список задач
```bash
GET /tasks

Query параметры (опционально):

status (например: created, in_progress, completed, expired)

responsible_id (ID пользователя)

limit (По умолчанию: 10)

offset (По умолчанию: 0)
```
Response (200): Массив объектов задач.

3. Получить задачу по ID
```bash
GET /tasks/{id}
```
Response (200): Объект задачи.

6. Обновить задачу
```bash
PUT /tasks/{id}
```
Ограничение: Только создатель задачи может её изменять (Иначе 403 Forbidden).

Body: Содержит обновленные поля задачи (title, description, status, deadline, responsible_id).

Response (200): Обновленный объект задачи.

7. Удалить задачу
```bash
DELETE /tasks/{id}
```
Ограничение: Только создатель задачи может её удалить (Иначе 403 Forbidden).

Response (204 No Content)

















