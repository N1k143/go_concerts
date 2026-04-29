# Concerts REST API

REST API для системы бронирования билетов на концерты. Написан на Go с PostgreSQL и Docker.

## Стек

- **Go 1.22** + chi router
- **PostgreSQL 15** (Docker)
- **sqlx** для работы с БД
- **Docker Compose** для запуска

---

## 🚀 Запуск

### Требования

- Docker и Docker Compose
- (опционально) Go 1.22+ для локальной разработки

### Запуск через Docker Compose

```bash
# 1. Клонировать / распаковать проект
cd concerts

# 2. Запустить все сервисы
docker compose up --build

# API будет доступен на http://localhost:8080
```

При первом запуске Docker автоматически:
- Создаст базу данных PostgreSQL
- Выполнит миграции (таблицы + тестовые данные)
- Соберёт и запустит Go-сервер

### Остановка

```bash
docker compose down          # остановить
docker compose down -v       # остановить и удалить данные
```

---

## 🧪 Тестирование API

Все примеры используют `curl`. Предполагается, что сервер запущен на `http://localhost:8080`.

---

### GET /api/v1/concerts — Список концертов

```bash
curl http://localhost:8080/api/v1/concerts
```

---

### GET /api/v1/concerts/{id} — Один концерт

```bash
# Существующий концерт
curl http://localhost:8080/api/v1/concerts/1

# Несуществующий (вернёт 404)
curl http://localhost:8080/api/v1/concerts/999
```

---

### GET /api/v1/concerts/{id}/shows/{show-id}/seating — Схема зала

```bash
# Шоу 1 концерта 1
curl http://localhost:8080/api/v1/concerts/1/shows/1/seating

# Неверная пара (шоу не принадлежит концерту) — 404
curl http://localhost:8080/api/v1/concerts/1/shows/2/seating
```

---

### POST /api/v1/concerts/{id}/shows/{show-id}/reservation — Резервация мест

```bash
# Зарезервировать места (ряд id=1, места 10 и 11)
curl -s -X POST http://localhost:8080/api/v1/concerts/1/shows/1/reservation \
  -H "Content-Type: application/json" \
  -d '{
    "reservations": [
      {"row": 3, "seat": 5},
      {"row": 3, "seat": 6}
    ],
    "duration": 120
  }'
# Ответ содержит reservation_token — сохраните его!
```

```bash
# Заменить резервацию (использовать предыдущий токен)
curl -s -X POST http://localhost:8080/api/v1/concerts/1/shows/1/reservation \
  -H "Content-Type: application/json" \
  -d '{
    "reservation_token": "ВАШ_ТОКЕН_ЗДЕСЬ",
    "reservations": [
      {"row": 3, "seat": 7}
    ],
    "duration": 300
  }'
```

```bash
# Ошибка: место уже занято
curl -s -X POST http://localhost:8080/api/v1/concerts/1/shows/1/reservation \
  -H "Content-Type: application/json" \
  -d '{
    "reservations": [{"row": 1, "seat": 1}]
  }'
```

---

### POST /api/v1/concerts/{id}/shows/{show-id}/booking — Оформить бронирование

```bash
# Сначала зарезервируйте места (см. выше) и получите TOKEN
curl -s -X POST http://localhost:8080/api/v1/concerts/1/shows/1/booking \
  -H "Content-Type: application/json" \
  -d '{
    "reservation_token": "ВАШ_ТОКЕН_ЗДЕСЬ",
    "name": "Ivan Petrov",
    "address": "Lenina 15",
    "city": "Astana",
    "zip": "010000",
    "country": "Kazakhstan"
  }'
# Ответ содержит список билетов с кодами — сохраните код билета!
```

---

### POST /api/v1/tickets — Получить билеты по коду

```bash
curl -s -X POST http://localhost:8080/api/v1/tickets \
  -H "Content-Type: application/json" \
  -d '{
    "code": "КОД_БИЛЕТА_ЗДЕСЬ",
    "name": "Ivan Petrov"
  }'
```

---

### POST /api/v1/tickets/{ticket-id}/cancel — Отменить билет

```bash
curl -s -X POST http://localhost:8080/api/v1/tickets/1/cancel \
  -H "Content-Type: application/json" \
  -d '{
    "code": "КОД_БИЛЕТА_ЗДЕСЬ",
    "name": "Ivan Petrov"
  }'
# Успех: HTTP 204 No Content
```

---

## 📋 Полный сценарий тестирования

```bash
# 1. Посмотреть концерты
curl -s http://localhost:8080/api/v1/concerts | jq .

# 2. Посмотреть схему зала шоу 1 концерта 1
curl -s http://localhost:8080/api/v1/concerts/1/shows/1/seating | jq '.rows[0]'

# 3. Зарезервировать места
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/concerts/1/shows/1/reservation \
  -H "Content-Type: application/json" \
  -d '{"reservations":[{"row":5,"seat":1},{"row":5,"seat":2}],"duration":300}' \
  | jq -r '.reservation_token')

echo "Token: $TOKEN"

# 4. Оформить бронирование
TICKET_CODE=$(curl -s -X POST http://localhost:8080/api/v1/concerts/1/shows/1/booking \
  -H "Content-Type: application/json" \
  -d "{\"reservation_token\":\"$TOKEN\",\"name\":\"Test User\",\"address\":\"Test St 1\",\"city\":\"Almaty\",\"zip\":\"050000\",\"country\":\"Kazakhstan\"}" \
  | jq -r '.tickets[0].code')

echo "Ticket code: $TICKET_CODE"

# 5. Получить все билеты по коду
curl -s -X POST http://localhost:8080/api/v1/tickets \
  -H "Content-Type: application/json" \
  -d "{\"code\":\"$TICKET_CODE\",\"name\":\"Test User\"}" | jq .

# 6. Получить ID первого билета и отменить его
TICKET_ID=$(curl -s -X POST http://localhost:8080/api/v1/tickets \
  -H "Content-Type: application/json" \
  -d "{\"code\":\"$TICKET_CODE\",\"name\":\"Test User\"}" | jq -r '.tickets[0].id')

curl -s -X POST http://localhost:8080/api/v1/tickets/$TICKET_ID/cancel \
  -H "Content-Type: application/json" \
  -d "{\"code\":\"$TICKET_CODE\",\"name\":\"Test User\"}"
# Должно вернуть HTTP 204
```

---

## Структура проекта

```
concerts/
├── cmd/api/main.go              # Точка входа
├── internal/
│   ├── config/config.go         # Конфигурация из env
│   ├── database/db.go           # Подключение к БД
│   ├── handlers/                # HTTP обработчики
│   │   ├── concert.go           # GET /concerts, GET /concerts/{id}
│   │   ├── seating.go           # GET .../seating
│   │   ├── reservation.go       # POST .../reservation
│   │   ├── booking.go           # POST .../booking, tickets
│   │   └── helpers.go           # writeJSON, writeError
│   ├── middleware/logger.go     # Цветной логгер запросов
│   ├── models/models.go         # Структуры данных
│   └── repository/              # Слой доступа к БД
│       ├── concert.go
│       ├── seating.go
│       ├── reservation.go
│       └── booking.go
├── migrations/001_init.sql      # PostgreSQL схема + seed данные
├── docker-compose.yml
├── Dockerfile
├── go.mod / go.sum
└── .env                         # Для локального запуска
```

## API эндпоинты

| Метод | Путь | Описание |
|-------|------|----------|
| GET | /api/v1/concerts | Список всех концертов |
| GET | /api/v1/concerts/{id} | Один концерт |
| GET | /api/v1/concerts/{id}/shows/{id}/seating | Схема зала |
| POST | /api/v1/concerts/{id}/shows/{id}/reservation | Резервация |
| POST | /api/v1/concerts/{id}/shows/{id}/booking | Бронирование |
| POST | /api/v1/tickets | Получить билеты по коду |
| POST | /api/v1/tickets/{id}/cancel | Отменить билет |
