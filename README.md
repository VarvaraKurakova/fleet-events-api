# Fleet Events API

Fleet Events API — это backend-сервис на Go для приёма и обработки telemetry-событий от GPS/IoT-устройств, установленных на транспорт.

Идея проекта простая: устройство присылает событие с координатами, скоростью, зарядом батареи и другой телеметрией, API сохраняет это событие, обновляет последнее состояние машины, публикует сообщение в RabbitMQ, а отдельный worker асинхронно обрабатывает событие и создаёт alerts.

Проект сделан как pet-project для демонстрации backend-навыков: REST API, PostgreSQL, Redis, RabbitMQ, Docker Compose, graceful shutdown, middleware, worker pool и unit-тесты сервисного слоя.

---

## Что умеет сервис

- Управлять автопарками, машинами и устройствами.
- Принимать telemetry events от устройств через HTTP API.
- Хранить события в PostgreSQL.
- Кешировать последнее состояние машины в Redis.
- Публиковать событие `event.created` в RabbitMQ.
- Обрабатывать события отдельным worker-сервисом.
- Создавать alerts по простым бизнес-правилам.
- Возвращать историю событий машины с пагинацией.
- Возвращать alerts с фильтрами и пагинацией.

---

## Архитектура

```text
GPS / IoT Device
        |
        | HTTP JSON
        v
Go API Service
        |
        | save event
        v
PostgreSQL
        |
        | update latest vehicle state
        v
Redis
        |
        | publish event.created
        v
RabbitMQ
        |
        | consume messages
        v
Go Worker Service
        |
        | create alerts
        v
PostgreSQL
```

Основной поток выглядит так:

1. Устройство отправляет telemetry event в `POST /api/v1/events`.
2. API проверяет `X-API-Key` и валидирует входные данные.
3. `EventService` ищет устройство по `external_id`.
4. Событие сохраняется в PostgreSQL.
5. У устройства обновляется `last_seen_at`.
6. Последнее состояние машины сохраняется в Redis.
7. В RabbitMQ публикуется сообщение `event.created`.
8. Worker получает сообщение из очереди.
9. Worker проверяет событие по alert-правилам.
10. Если есть нарушения, alerts сохраняются в PostgreSQL.

---

## Технологии

- Go
- chi router
- PostgreSQL
- pgx / pgxpool
- Redis
- RabbitMQ
- slog
- Docker / Docker Compose
- golang-migrate
- Unit tests

---

## Структура проекта

```text
.
├── cmd/
│   ├── api/              # запуск HTTP API
│   └── worker/           # запуск worker-сервиса
├── internal/
│   ├── app/              # сборка приложения
│   ├── config/           # конфигурация из env
│   ├── domain/           # доменные модели
│   ├── http/             # router, handlers, middleware
│   ├── service/          # бизнес-логика
│   ├── repository/       # PostgreSQL и Redis repository/cache
│   ├── messaging/        # сообщения и RabbitMQ
│   ├── worker/           # worker pool
│   └── apperrors/        # общие ошибки приложения
├── migrations/           # SQL migrations
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── README.md
```

---

## Alert-правила

Worker создаёт alerts по данным из telemetry event.

| Условие | Alert type |
|---|---|
| `speed > 90` | `speed_limit_exceeded` |
| `battery_level < 0.15` | `low_battery` |
| нет координат или координаты невалидны | `invalid_location` |

Сейчас это простые правила для MVP. В production-версии их можно было бы вынести в отдельный rule engine или сделать конфигурируемыми для разных fleet/vehicle.

---

## API endpoints

### Health

```http
GET /health
GET /ready
```

`/health` проверяет, что API живой.

`/ready` проверяет готовность зависимостей: PostgreSQL, Redis и RabbitMQ.

---

### Fleets

```http
POST /api/v1/fleets
GET  /api/v1/fleets
GET  /api/v1/fleets/{id}
```

---

### Vehicles

```http
POST /api/v1/vehicles
GET  /api/v1/vehicles
GET  /api/v1/vehicles/{id}
GET  /api/v1/vehicles/{id}/state
GET  /api/v1/vehicles/{id}/events?limit=50&offset=0
```

`GET /api/v1/vehicles/{id}/state` сначала пытается взять последнее состояние машины из Redis. Если в Redis ничего нет, сервис берёт последнее событие из PostgreSQL, собирает из него state, кладёт state обратно в Redis и возвращает ответ.

---

### Devices

```http
POST /api/v1/devices
GET  /api/v1/devices
GET  /api/v1/devices/{id}
```

---

### Events

```http
POST /api/v1/events
```

Для отправки события нужен API key header:

```http
X-API-Key: dev-api-key
```

---

### Alerts

```http
GET   /api/v1/alerts
GET   /api/v1/alerts?status=open&type=low_battery&limit=50&offset=0
PATCH /api/v1/alerts/{id}/resolve
```

Поддерживаемые query params для alerts:

```text
status
type
vehicle_id
limit
offset
```

---

## Переменные окружения

Пример `.env`:

```env
APP_ENV=local
HTTP_ADDR=:8080

POSTGRES_DB=fleet_events
POSTGRES_USER=fleet
POSTGRES_PASSWORD=fleet
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DSN=postgres://fleet:fleet@localhost:5432/fleet_events?sslmode=disable

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

RABBITMQ_URL=amqp://guest:guest@localhost:5672/

DEVICE_API_KEY=dev-api-key

WORKER_CONCURRENCY=4
WORKER_PREFETCH=4
```

Для локального запуска API с Mac используется `localhost`, потому что PostgreSQL/Redis/RabbitMQ проброшены наружу через Docker ports.

Внутри `docker-compose.yml` сервисы обращаются друг к другу по именам контейнеров: `postgres`, `redis`, `rabbitmq`.

---

## Запуск локально

Сначала поднимаем инфраструктуру:

```bash
docker compose up -d postgres redis rabbitmq
```

Применяем миграции:

```bash
make migrate-up
```

Запускаем API:

```bash
make run-api
```

В отдельном терминале запускаем worker:

```bash
make run-worker
```

---

## Запуск через Docker Compose

Запустить всё сразу:

```bash
docker compose up --build
```

Остановить сервисы:

```bash
docker compose down
```

Посмотреть логи API:

```bash
docker compose logs -f api
```

Посмотреть логи worker:

```bash
docker compose logs -f worker
```

---

## Миграции

Применить миграции:

```bash
make migrate-up
```

Откатить последнюю миграцию:

```bash
make migrate-down
```

Посмотреть текущую версию миграций:

```bash
make migrate-version
```

---

## Примеры запросов

### Health check

```bash
curl -i http://localhost:8080/health
```

### Readiness check

```bash
curl -i http://localhost:8080/ready
```

---

### Создать fleet

```bash
curl -X POST http://localhost:8080/api/v1/fleets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Moscow Delivery Fleet"
  }'
```

---

### Создать vehicle

```bash
curl -X POST http://localhost:8080/api/v1/vehicles \
  -H "Content-Type: application/json" \
  -d '{
    "fleet_id": "PASTE_FLEET_ID",
    "plate_number": "A001AA777",
    "model": "Ford Transit"
  }'
```

---

### Создать device

```bash
curl -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -d '{
    "vehicle_id": "PASTE_VEHICLE_ID",
    "external_id": "tracker-100500",
    "model": "Teltonika FMB920"
  }'
```

---

### Отправить telemetry event

Этот пример создаст сразу два alerts: превышение скорости и низкий заряд батареи.

```bash
curl -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-api-key" \
  -d '{
    "device_id": "tracker-100500",
    "event_type": "telemetry",
    "event_time": "2026-05-11T10:15:00Z",
    "location": {
      "lat": 55.7558,
      "lon": 37.6173
    },
    "speed": 110.5,
    "battery_level": 0.10,
    "ignition": true,
    "metadata": {
      "source": "test"
    }
  }'
```

---

### Получить последнее состояние машины

```bash
curl http://localhost:8080/api/v1/vehicles/PASTE_VEHICLE_ID/state
```

---

### Получить события машины

```bash
curl "http://localhost:8080/api/v1/vehicles/PASTE_VEHICLE_ID/events?limit=10&offset=0"
```

---

### Получить alerts

```bash
curl "http://localhost:8080/api/v1/alerts?limit=10&offset=0"
```

---

### Получить открытые alerts по низкой батарее

```bash
curl "http://localhost:8080/api/v1/alerts?status=open&type=low_battery&limit=10&offset=0"
```

---

### Закрыть alert

```bash
curl -X PATCH http://localhost:8080/api/v1/alerts/PASTE_ALERT_ID/resolve
```

---

## Тесты

Запустить все тесты:

```bash
go test ./...
```

Запустить только тесты сервисного слоя:

```bash
go test ./internal/service
```

Сейчас unit-тестами покрыты:

- `AlertService`: правила создания alerts.
- `VehicleStateService`: Redis hit и Redis miss fallback в PostgreSQL.
- `EventService`: основной ingestion flow.

---
