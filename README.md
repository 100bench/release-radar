# ReleaseRadar

ReleaseRadar — это трекер релизов GitHub с уведомлениями в Telegram.

## Возможности

-   **Отслеживание релизов GitHub:** Опрос GitHub для получения новых релизов с использованием ETag для эффективности.
-   **Уведомления в Telegram:** Отправка уведомлений о релизах в настроенные каналы Telegram.
-   **Чистая архитектура:** Проект организован по принципам чистой архитектуры (слои `internal/domain`, `internal/usecase`, `internal/adapter`).
-   **Два бинарника:** `api` для обработки HTTP-запросов и `worker` для фоновых задач (опрос и уведомления).
-   **Постоянство данных:** PostgreSQL (через GORM) для хранения данных и Redis для идемпотентности и кэширования.
-   **Наблюдаемость:** Интеграция с Prometheus для метрик и OpenTelemetry для трассировки.
-   **Конфигурация:** Управление настройками через переменные окружения (Viper).
-   **Структурированное логирование:** Использование Zap для структурированных логов.
-   **Идемпотентность и повторные попытки:** Обеспечение надежной доставки сообщений и вызовов API.
-   **RESTful API:** Создан с использованием фреймворка Gin, включает спецификацию OpenAPI (Swagger UI).

## Структура проекта

```
.github/
├── workflows/
    └── ci.yml
cmd/
├── api/
│   └── main.go         # Точка входа API-сервиса
└── worker/
    └── main.go       # Точка входа Worker-сервиса
internal/
├── adapter/
│   ├── cache/
│   │   └── redis.go    # Реализация кэша на Redis
│   ├── github/
│   │   ├── client.go   # Интерфейс GitHub-клиента
│   │   ├── github.go   # Реализация GitHub-клиента
│   │   └── mock.go     # Мок GitHub-клиента для тестирования
│   ├── http/
│   │   └── handler.go  # HTTP-обработчики для API
│   ├── persistence/
│   │   ├── errors.go   # Пользовательские ошибки слоя персистентности
│   │   ├── postgres.go # Реализация PostgreSQL через GORM
│   │   ├── redis.go    # Реализация хранилища идемпотентности на Redis
│   │   └── repository.go # Интерфейсы репозиториев
│   └── telegram/
│       ├── client.go   # Интерфейс Telegram-клиента
│       ├── mock.go     # Мок Telegram-клиента для тестирования
│       └── telegram.go # Реализация Telegram-клиента
├── domain/
│   └── models.go       # Основные сущности/модели данных
└── usecase/
    ├── notifier.go     # Реализация варианта использования для уведомлений
    ├── poller.go       # Реализация варианта использования для опроса релизов
    ├── repo.go         # Реализация варианта использования для управления репозиториями
    ├── subscription.go # Реализация варианта использования для управления подписками
    └── usecase.go      # Интерфейсы вариантов использования
migrations/
└── 000001_initial_schema.up.sql # Начальная миграция схемы базы данных
pkg/
├── idempotency/
│   └── idempotency.go  # Менеджер идемпотентности
├── logger/
│   └── logger.go       # Структурированный логгер (Zap)
├── otel/
│   └── otel.go         # Заглушки для трассировки OpenTelemetry
└── retry/
    └── retry.go        # Механизм повторных попыток с экспоненциальной задержкой
build/
├── Dockerfile.api      # Dockerfile для API-сервиса
└── Dockerfile.worker   # Dockerfile для Worker-сервиса
docs/
├── docs.go             # Сгенерированный код документации Swagger
├── swagger.json
└── swagger.yaml
docker-compose.yml      # Конфигурация Docker Compose
Makefile                # Команды для сборки, тестирования, запуска и работы с Docker
openapi.yaml            # Спецификация OpenAPI 3.0
prometheus.yml          # Конфигурация Prometheus
seed.sql                # SQL-данные для инициализации
.env.example            # Пример переменных окружения
go.mod
go.sum
README.md
```

## Быстрый старт

Чтобы быстро запустить ReleaseRadar для разработки:

1.  **Клонируйте репозиторий:**

    ```bash
    git clone https://github.com/USER/releaseradar.git
    cd releaseradar
    ```

2.  **Установите `golangci-lint` и `swag`:**

    ```bash
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install github.com/swaggo/swag/cmd/swag@latest
    ```

3.  **Сгенерируйте документацию Swagger:**

    ```bash
    swag init -g cmd/api/main.go
    ```

4.  **Скопируйте пример переменных окружения:**

    ```bash
    cp .env.example .env
    # Отредактируйте .env, добавив ваш GitHub Token и Telegram Bot Token
    ```

5.  **Запустите сервисы с Docker Compose:**

    ```bash
    docker-compose up --build -d
    ```

    Это запустит PostgreSQL, Redis, API-сервис, Worker-сервис и Prometheus.

6.  **Доступ к API:**

    -   **API:** `http://localhost:8080/healthz`
    -   **Swagger UI:** `http://localhost:8080/swagger/index.html`
    -   **Prometheus:** `http://localhost:9090`

## Конфигурация

ReleaseRadar использует переменные окружения для конфигурации. См. `.env.example` для списка настраиваемых параметров.

## Разработка

-   **Сборка:** `make build`
-   **Тестирование:** `make test`
-   **Линтинг:** `make lint`
-   **Локальный запуск:** `make run` (запускает API и Worker в фоновом режиме)

## Лицензия

Этот проект лицензирован по лицензии Apache 2.0.
