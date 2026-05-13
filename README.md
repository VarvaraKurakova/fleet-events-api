# Fleet Events API

Backend-сервис на Go для приёма и обработки событий от транспортных устройств: GPS-трекеров, IoT-датчиков и похожих источников телеметрии.

Идея проекта простая: устройство отправляет событие, API принимает его, сохраняет в PostgreSQL, обновляет последнее состояние транспорта в Redis и отправляет сообщение в RabbitMQ. Отдельный worker потом обрабатывает событие и создаёт alerts, например при превышении скорости или низком заряде устройства.

Проект делается как pet-project для практики Go backend-разработки: HTTP API, работа с БД, Redis, RabbitMQ, фоновые workers, goroutines, graceful shutdown, тесты и Docker Compose.

## Current status

Сейчас реализовано:

- базовая структура Go-проекта;
- HTTP API service;
- endpoint `GET /health`;
- graceful shutdown для API;
- Docker Compose с PostgreSQL, Redis и RabbitMQ;
- базовый Makefile для запуска проекта.

## Tech stack

- Go
- net/http
- chi
- PostgreSQL
- Redis
- RabbitMQ
- Docker Compose

## Architecture idea

Планируемая архитектура:

```text
GPS / IoT device
        |
        | HTTP JSON
        v
Go API service
        |
        | saves event
        v
PostgreSQL
        |
        | updates last known state
        v
Redis
        |
        | publishes event.created
        v
RabbitMQ
        |
        | consumes messages
        v
Go worker service
        |
        | creates alerts
        v
PostgreSQL

## Локальный запуск
#Start infrastructure:

```bash
make up
```

Run API:

```bash
make run-api
```

Check health endpoint:

```bash
curl http://localhost:8080/health
```
