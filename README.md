# Go Payment Platform

Backend платежной платформы, разработанный на Go.

Проект реализует REST API для управления пользователями, аккаунтами и денежными переводами с JWT-аутентификацией и PostgreSQL.

## Stack

- Go
- PostgreSQL
- pgx
- JWT
- Docker / Docker Compose

## Features

- Регистрация пользователей
- Аутентификация
- JWT access tokens
- Refresh tokens
- Refresh token rotation
- Управление пользовательскими сессиями
- Logout с отзывом сессии
- Защита API через JWT middleware
- Создание и получение аккаунтов
- Пополнение баланса
- Денежные переводы
- История транзакций
- Проверка доступа к ресурсам пользователя
- Логирование HTTP-запросов
- PostgreSQL migrations

## Authentication

После успешного входа сервер создаёт:

- `access_token` — JWT для доступа к защищённым API
- `refresh_token` — для обновления access token

Токены передаются через `HttpOnly` cookies.

### Refresh Token Rotation

При обновлении токена:

1. Старый refresh token помечается как использованный.
2. Создаётся новый refresh token.
3. Генерируется новый access token.

Повторное использование refresh token предотвращается на уровне базы данных.

### Sessions

Каждая пользовательская сессия хранится отдельно.

Logout отзывает текущую сессию, после чего связанный refresh token больше не может использоваться.

## Architecture

```text
HTTP Request
     │
     ▼
Router
     │
     ▼
Middleware
(Auth / Logging)
     │
     ▼
Handlers
     │
     ▼
Service
(Business Logic)
     │
     ▼
Storage
     │
     ▼
PostgreSQL
```

## Protected Routes

Защищённые маршруты используют JWT middleware.

```text
Request
   │
   ▼
access_token cookie
   │
   ▼
JWT validation
   │
   ▼
userID → context
   │
   ▼
Handler
   │
   ▼
Service
   │
   ▼
Access validation
```

Пользователь может работать только с ресурсами, которые принадлежат ему.

## API

### Public

```text
GET  /health

POST /register
POST /login
POST /refresh
POST /logout
```

### Protected

```text
GET  /accounts/{id}

POST /deposits
POST /transfers

GET  /accounts/{id}/transactions
```

## Project Structure

```text
.
├── cmd/
│   └── api/
├── internal/
│   ├── auth/
│   ├── handler/
│   ├── httpapi/
│   ├── middleware/
│   ├── payment/
│   └── storage/
├── migrations/
├── docs/
├── compose.yaml
├── go.mod
└── go.sum
```
