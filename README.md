# Release-Radar

[![Go](https://img.shields.io/github/go-mod/go-version/100bench/release-radar?style=flat-square)](https://go.dev/)
[![License](https://img.shields.io/github/license/100bench/release-radar?style=flat-square)](LICENSE)
[![CI Status](https://img.shields.io/github/actions/workflow/status/100bench/release-radar/ci.yml?branch=main&style=flat-square)](https://github.com/100bench/release-radar/actions/workflows/ci.yml)

Release-Radar — трекер релизов GitHub → уведомления в Telegram.
Стек: Go, Gin, Postgres, Redis, Docker, Prometheus, OTEL. Worker+API, ETag, идемпотентность.

**Быстрый старт:**
```bash
cp .env.example .env && docker-compose up --build -d
```
Через 10–20 сек сервисы будут готовы; проверьте `/healthz`.
**API:** `http://localhost:8080/swagger/index.html` | Health: `/healthz` | Metrics: `:8080/metrics`

**Клонировать:** `git clone https://github.com/100bench/release-radar.git && cd release-radar`

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
├── api/            # Точка входа API-сервиса
└── worker/         # Точка входа Worker-сервиса
internal/           # Внутренняя логика приложения
├── adapter/        # Внешние адаптеры (GitHub, Telegram, PostgreSQL, Redis)
├── domain/         # Основные сущности/модели данных
└── usecase/        # Варианты использования/бизнес-логика
migrations/         # Миграции базы данных
pkg/                # Общие утилиты и библиотеки
build/              # Dockerfile'ы для сборки сервисов
docs/               # Документация OpenAPI/Swagger
docker-compose.yml  # Конфигурация Docker Compose
Makefile            # Скрипты для сборки, тестирования и запуска
openapi.yaml        # Спецификация OpenAPI 3.0
.env.example        # Пример переменных окружения
```

## Быстрый старт

Чтобы быстро запустить Release-Radar для разработки:

1.  **Клонируйте репозиторий:**

    ```bash
    git clone https://github.com/100bench/release-radar.git
    cd release-radar
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
    <br/>
    *Примечание: `openapi.yaml` является основной, вручную поддерживаемой спецификацией OpenAPI. `swag init` используется для генерации документации Swagger UI из комментариев в коде, чтобы обеспечить визуализацию API.*

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
    Через 10–20 сек сервисы будут готовы; проверьте `/healthz`.

6.  **Доступ к API:**

    -   **API:** `http://localhost:8080/healthz`
    -   **Swagger UI:** `http://localhost:8080/swagger/index.html`
    -   **Prometheus:** `http://localhost:9090`

### Примеры API запросов:

```bash
# добавить репозиторий
curl -X POST http://localhost:8080/api/v1/repos -d '{"full_name":"golang/go"}' -H 'Content-Type: application/json'
# подписка на уведомления в Telegram
curl -X POST http://localhost:8080/api/v1/subscriptions -d '{"repo_id":1,"channels":["telegram"]}' -H 'Content-Type: application/json'
```

## Конфигурация

Ключевые переменные окружения:
*   `RR_GITHUB_TOKEN`: токен доступа GitHub (scopes: `public_repo`).
*   `RR_TELEGRAM_BOT_TOKEN`: токен вашего бота Telegram.
*   `RR_POSTGRES_DSN`: строка подключения к PostgreSQL.
*   `RR_REDIS_ADDR`: адрес сервера Redis.

Полный список настроек см. в файле `.env.example`.
Release-Radar использует переменные окружения для конфигурации. См. `.env.example` для списка настраиваемых параметров.

## Разработка

-   **Сборка:** `make build`
-   **Тестирование:** `make test`
-   **Линтинг:** `make lint`
-   **Локальный запуск:** `make run` (запускает API и Worker в фоновом режиме)
