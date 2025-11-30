# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Обзор проекта

SMC-SellerService - микросервис для управления компаниями (автомойками) и их услугами в платформе онлайн-записи на автомойку. Сервис работает на порту 8081 и предоставляет публичные и защищённые endpoints для CRUD операций.

**Статус проекта:** ✅ Протестировано и работает согласно OpenAPI спецификации

### Tech Stack

- **Language**: Go 1.24.2
- **Architecture**: Clean Architecture (Domain, Service, Repository, Handlers)
- **Database**: PostgreSQL 16 + lib/pq + golang-migrate
- **HTTP Router**: Gorilla Mux
- **Query Builder**: Squirrel (psqlbuilder wrapper)
- **Authentication**: Simplified (X-User-ID + X-User-Role headers for MVP)
- **Logging**: Custom logger (console + file, injectable dependency)
- **Metrics**: Prometheus (HTTP + Database metrics)
- **Monitoring**: Централизованный мониторинг через SMC-Monitoring
- **Containerization**: Docker Compose
- **Module**: `github.com/m04kA/SMC-SellerService`

## Development Commands

### Makefile команды (рекомендуется)

```bash
# Справка по всем командам
make help

# Запуск всех сервисов в Docker
make docker-up

# Просмотр логов приложения
make docker-logs-app

# Остановка всех сервисов
make docker-down

# Полная очистка (volumes + images)
make clean-all

# Разработка с локальным запуском
make dev          # Запустить только БД
make run          # Запустить приложение локально

# Сброс базы данных
make db-reset

# Загрузка тестовых фикстур
make fixtures-load
```

### Quick Start (Docker)

```bash
# Запуск всех сервисов
make docker-up
# или
docker-compose up -d

# Просмотр логов
make docker-logs
# или
docker-compose logs -f

# Остановка всех сервисов
make docker-down
# или
docker-compose down

# Полная очистка (volumes + images)
make clean-all
# или
docker-compose down -v
```

### Local Development

```bash
# Установка зависимостей (первый раз)
make install
# или
go mod tidy

# Запуск приложения локально
make run
# или
go run cmd/main.go

# Сборка бинарного файла
make build
# или
go build -o bin/smc-sellerservice cmd/main.go
```

### Database Management

```bash
# Применить миграции (автоматически через docker-compose)
docker-compose up -d postgres
docker-compose up migrate

# Откат миграции
docker-compose run --rm migrate -path /migrations -database "postgres://postgres:postgres@postgres:5432/smc_sellerservice?sslmode=disable" down

# Загрузка тестовых фикстур (рекомендуется)
make fixtures-load
# Загружает данные из migrations/fixtures/*.sql:
# - 3 компании (Автомойка Премиум, СТО Профи, Детейлинг Центр)
# - 4 адреса с координатами
# - 5 услуг с привязкой к адресам
# Используется ON CONFLICT для безопасного повторного применения
```

### Database

- PostgreSQL работает на порту **5436** (не стандартный 5432, чтобы избежать конфликтов)
- Connection string: `host=localhost port=5436 user=postgres password=postgres dbname=smc_sellerservice sslmode=disable`
- Миграции автоматически применяются при `docker-compose up` через контейнер `migrate`

### Testing

API полностью протестирован и соответствует OpenAPI спецификации:

```bash
# Тестирование с помощью curl
# См. подробные команды в test_data/API_TEST_COMMANDS.md

# Тестирование с помощью Bruno
# Коллекция находится в /Users/yapanarin/GolandProjects/SMC-Bruno/SMC/SellerService
# 16 готовых запросов для всех endpoints
```

### Monitoring and Metrics

Сервис экспортирует метрики в формате Prometheus на `/metrics` endpoint:

```bash
# Просмотр метрик
curl http://localhost:8081/metrics
```

**Централизованный мониторинг:**
- Метрики автоматически собираются сервисом **SMC-Monitoring** (https://github.com/m04kA/SMC-Monitoring)
- **Prometheus** scrapes метрики каждые 15 секунд с endpoint `host.docker.internal:8081/metrics`
- **Grafana** визуализирует метрики на дашборде "SMC-SellerService Metrics"
- **PostgreSQL Exporter** собирает метрики базы данных

**Доступные метрики:**
- `http_requests_total` - количество HTTP запросов (labels: method, endpoint, status_code, service)
- `http_request_duration_seconds` - длительность HTTP запросов (histogram)
- `db_queries_total` - количество SQL запросов (labels: operation, table, status, service)
- `db_query_duration_seconds` - длительность SQL запросов (histogram)
- `db_connections_active` - активные соединения с БД (gauge)
- `db_connections_idle` - простаивающие соединения в пуле (gauge)
- `db_connections_max` - максимум соединений (gauge)

**Конфигурация метрик:**
```toml
# config.toml
[metrics]
enabled = true
path = "/metrics"
service_name = "sellerservice"
```

Переопределение через env:
```bash
METRICS_ENABLED=true
METRICS_PATH=/metrics
METRICS_SERVICE_NAME=sellerservice
```

**Ключевые пакеты для метрик:**
- `pkg/dbmetrics` - обёртка над sql.DB для сбора метрик БД
- `pkg/dbmetrics/interfaces.go` - интерфейсы DBExecutor, TxExecutor, SqlTxWrapper
- `internal/api/middleware/metrics.go` - middleware для HTTP метрик (если используется)

## Архитектура

Проект следует принципам **Clean Architecture** с чётким разделением ответственности:

```
SMC-SellerService/
├── cmd/
│   └── main.go                          # Entry point с routing и DI
├── internal/
│   ├── config/
│   │   └── config.go                   # Config loader (TOML)
│   ├── domain/                          # Доменные модели (чистые Go структуры)
│   │   ├── company.go
│   │   ├── address.go
│   │   ├── working_hours.go
│   │   └── service.go
│   ├── infra/storage/                   # Реализация репозиториев (PostgreSQL)
│   │   ├── company/
│   │   │   ├── contract.go             # Интерфейс DBExecutor
│   │   │   ├── errors.go               # Sentinel errors репозитория
│   │   │   └── repository.go           # CRUD + связанные сущности
│   │   └── service/
│   │       ├── contract.go
│   │       ├── errors.go
│   │       └── repository.go
│   ├── service/                         # Бизнес-логика и DTO модели
│   │   ├── constants.go                # RoleSuperuser, RoleUser
│   │   ├── companies/
│   │   │   ├── contracts.go            # Интерфейсы зависимостей
│   │   │   ├── errors.go               # Service-level errors
│   │   │   ├── service.go              # Бизнес-логика + авторизация
│   │   │   └── models/
│   │   │       └── models.go           # DTOs и конвертеры
│   │   └── services/
│   │       ├── contracts.go
│   │       ├── errors.go
│   │       ├── service.go
│   │       └── models/
│   │           └── models.go
│   ├── api/                             # HTTP слой
│   │   ├── handlers/
│   │   │   ├── utils.go                # RespondJSON, RespondError, DecodeJSON
│   │   │   ├── create_company/
│   │   │   │   ├── contract.go         # Интерфейсы (Service, Logger)
│   │   │   │   └── handler.go          # HTTP handler
│   │   │   ├── get_company/
│   │   │   ├── list_companies/
│   │   │   ├── update_company/
│   │   │   ├── delete_company/
│   │   │   ├── create_service/
│   │   │   ├── get_service/
│   │   │   ├── list_services/
│   │   │   ├── update_service/
│   │   │   └── delete_service/
│   │   └── middleware/
│   │       └── auth.go                 # UserIDAuth middleware
│   └── integrations/                    # Клиенты для внешних сервисов
│       └── priceservice/                # Клиент для SMC-PriceService
│           ├── client.go                # HTTP клиент с GetPrice()
│           ├── contract.go              # Интерфейс Logger
│           ├── errors.go                # Sentinel errors
│           └── models.go                # PriceResponse DTO
├── pkg/
│   ├── logger/
│   │   └── logger.go                   # Injectable logger (Info/Warn/Error)
│   ├── dbmetrics/                       # Prometheus метрики для БД
│   │   ├── db.go                       # Обёртка над sql.DB с метриками
│   │   ├── interfaces.go               # DBExecutor, TxExecutor, SqlTxWrapper
│   │   └── metrics.go                  # Prometheus collectors
│   └── psqlbuilder/
│       └── psqlbuilder.go              # Обёртка над squirrel
├── migrations/
│   ├── 000001_init_schema.up.sql
│   ├── 000001_init_schema.down.sql
│   └── ...
├── test_data/                           # Тестовые данные и команды
│   ├── API_TEST_COMMANDS.md            # Набор curl команд для тестирования
│   ├── create_company.json
│   ├── update_company.json
│   ├── create_service.json
│   └── update_service.json
├── schemas/
│   └── schema.yaml                      # OpenAPI 3.0 спецификация
├── config.toml                          # Конфигурация приложения
├── docker-compose.yml                   # Docker окружение
├── Dockerfile.sellerservice             # Multi-stage Dockerfile
├── Makefile                             # Команды для разработки
├── .env                                 # Переменные окружения (не в git)
├── .env.example                         # Пример .env файла
└── go.mod
```

### Ключевые архитектурные паттерны

#### 1. Чистый Domain слой

Доменные модели (`internal/domain/`) - чистые Go структуры, представляющие бизнес-сущности.

**Важно:** Domain слой может содержать sql.Scanner и driver.Valuer для кастомных типов (например, TimeString для работы с PostgreSQL TIME полями), но НЕ содержит бизнес-логику.

**Пример кастомного типа:**
```go
// domain/working_hours.go
type TimeString string

// Scan implements sql.Scanner - для чтения из БД
func (t *TimeString) Scan(value interface{}) error {
    if value == nil {
        return nil
    }
    switch v := value.(type) {
    case time.Time:
        *t = TimeString(v.Format("15:04")) // PostgreSQL TIME -> "HH:MM"
        return nil
    case string:
        *t = TimeString(v)
        return nil
    default:
        return fmt.Errorf("cannot scan type %T into TimeString", value)
    }
}

// Value implements driver.Valuer - для записи в БД
func (t TimeString) Value() (driver.Value, error) {
    if t == "" {
        return nil, nil
    }
    return string(t), nil
}
```

#### 2. Repository паттерн

Слой хранения (`internal/infra/storage/`) реализует репозитории:

**Структура репозитория:**
- `contract.go` - интерфейс `DBExecutor` (для подмены *sql.DB/*sql.Tx)
- `errors.go` - sentinel errors (ErrCompanyNotFound, ErrBuildQuery, ErrExecQuery, ErrScanRow, ErrTransaction)
- `repository.go` - CRUD методы

**Обработка ошибок в репозитории:**
```go
// Используй fmt.Errorf с %w для wrapping
if err == sql.ErrNoRows {
    return nil, ErrCompanyNotFound
}
if err != nil {
    return nil, fmt.Errorf("%w: GetByID - scan company: %v", ErrScanRow, err)
}
```

**Импорт в сервисном слое:**
```go
import (
    companyRepo "github.com/m04kA/SMC-SellerService/internal/infra/storage/company"
)

// Проверка ошибок через errors.Is()
if errors.Is(err, companyRepo.ErrCompanyNotFound) {
    return nil, ErrCompanyNotFound
}
```

#### 3. Service слой с DTO

Слой бизнес-логики (`internal/service/`) содержит:

**Структура сервиса:**
- `contracts.go` - интерфейсы репозиториев
- `errors.go` - service-level errors (ErrCompanyNotFound, ErrAccessDenied, ErrInternal)
- `service.go` - бизнес-логика с авторизацией
- `models/models.go` - DTOs и конвертеры (ToDomain*, FromDomain*)

**Проверка авторизации:**
```go
func (s *Service) Update(ctx context.Context, userID int64, userRole string, id int64, req *models.UpdateCompanyRequest) (*models.CompanyResponse, error) {
    // Superuser может всё
    if userRole == constants.RoleSuperuser {
        // ... обновление
    }

    // User может только свои компании
    isManager, err := s.companyRepo.IsManager(ctx, id, userID)
    if err != nil {
        return nil, fmt.Errorf("%w: IsManager check failed: %v", ErrInternal, err)
    }
    if !isManager {
        return nil, ErrAccessDenied
    }
    // ... обновление
}
```

#### 4. Handler per Endpoint паттерн

Каждый endpoint в отдельном пакете (`internal/api/handlers/`):

**Структура handler:**
- `contract.go` - интерфейсы зависимостей (Service, Logger)
- `handler.go` - HTTP обработчик с константами для сообщений

**Пример handler.go:**
```go
package create_company

import (
    "errors"
    "net/http"
    "github.com/m04kA/SMC-SellerService/internal/api/handlers"
    "github.com/m04kA/SMC-SellerService/internal/api/middleware"
    "github.com/m04kA/SMC-SellerService/internal/service/companies"
    "github.com/m04kA/SMC-SellerService/internal/service/companies/models"
)

const (
    msgInvalidRequestBody = "invalid request body"
    msgForbidden          = "access denied"
    msgMissingUserID      = "missing user ID"
    msgMissingUserRole    = "missing user role"
)

type Handler struct {
    service CompanyService
    logger  Logger
}

func NewHandler(service CompanyService, logger Logger) *Handler {
    return &Handler{
        service: service,
        logger:  logger,
    }
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
    // Извлечение userID и userRole из контекста
    userID, ok := middleware.GetUserID(r.Context())
    if !ok {
        handlers.RespondUnauthorized(w, msgMissingUserID)
        return
    }

    userRole, ok := middleware.GetUserRole(r.Context())
    if !ok {
        handlers.RespondUnauthorized(w, msgMissingUserRole)
        return
    }

    // Парсинг request body
    var req models.CreateCompanyRequest
    if err := handlers.DecodeJSON(r, &req); err != nil {
        h.logger.Warn("POST /companies - Invalid request body: %v", err)
        handlers.RespondBadRequest(w, msgInvalidRequestBody)
        return
    }

    // Вызов сервисного слоя
    company, err := h.service.Create(r.Context(), userID, userRole, &req)
    if err != nil {
        // Проверка через errors.Is()
        if errors.Is(err, companies.ErrOnlySuperuser) {
            h.logger.Warn("POST /companies - Access denied: user_id=%d, role=%s", userID, userRole)
            handlers.RespondForbidden(w, msgForbidden)
            return
        }
        h.logger.Error("POST /companies - Failed: user_id=%d, error=%v", userID, err)
        handlers.RespondInternalError(w)
        return
    }

    h.logger.Info("POST /companies - Created: company_id=%d, user_id=%d", company.ID, userID)
    handlers.RespondJSON(w, http.StatusCreated, company)
}
```

**Общие утилиты (`handlers/utils.go`):**
- `RespondJSON(w, status, payload)` - JSON ответ
- `RespondError(w, status, message)` - ошибка с `ErrorResponse{Code, Message}`
- `DecodeJSON(r, v)` - парсинг body
- `RespondBadRequest(w, msg)` - 400
- `RespondUnauthorized(w, msg)` - 401
- `RespondForbidden(w, msg)` - 403
- `RespondNotFound(w, msg)` - 404
- `RespondInternalError(w)` - 500

#### 5. Injectable Logger

Logger передаётся через конструктор, описан как интерфейс в `contract.go` каждого handler:

```go
// contract.go
type Logger interface {
    Info(format string, v ...interface{})
    Warn(format string, v ...interface{})
    Error(format string, v ...interface{})
}

// Инициализация в main.go
log, err := logger.New("./logs/app.log")
if err != nil {
    panic(err)
}
defer log.Close()
```

**Логирование:**
- **INFO** - успешные операции (только консоль)
- **WARN** - 4xx ошибки (консоль + файл `logs/app.log`)
- **ERROR** - 5xx ошибки (консоль + файл `logs/app.log`)

#### 6. Построение SQL запросов

Используй кастомную обёртку `pkg/psqlbuilder` (обёртка над `squirrel`):

```go
import (
    "github.com/m04kA/SMC-SellerService/pkg/psqlbuilder"
    "github.com/Masterminds/squirrel"
)

// SELECT
query, args, err := psqlbuilder.Select("id", "name").
    From("companies").
    Where(squirrel.Eq{"id": id}).
    ToSql()

// INSERT
query, args, err := psqlbuilder.Insert("companies").
    Columns("name", "description").
    Values(name, description).
    Suffix("RETURNING id").
    ToSql()

// UPDATE
query, args, err := psqlbuilder.Update("companies").
    Set("name", name).
    Where(squirrel.Eq{"id": id}).
    ToSql()

// DELETE
query, args, err := psqlbuilder.Delete("companies").
    Where(squirrel.Eq{"id": id}).
    ToSql()
```

**НЕ используй плейсхолдеры напрямую:**
```go
// ❌ Неправильно
Where("id = ?", id)

// ✅ Правильно
Where(squirrel.Eq{"id": id})
```

#### 7. Управление транзакциями

```go
// Начало транзакции
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return fmt.Errorf("%w: BeginTx: %v", ErrTransaction, err)
}

// Явный rollback при ошибках (НЕ используй defer)
company, err := r.createCompany(ctx, tx, req)
if err != nil {
    tx.Rollback()
    return nil, err
}

addresses, err := r.createAddresses(ctx, tx, companyID, req.Addresses)
if err != nil {
    tx.Rollback()
    return nil, err
}

// Commit только после успеха
if err := tx.Commit(); err != nil {
    return nil, fmt.Errorf("%w: Commit: %v", ErrTransaction, err)
}
```

#### 8. Интеграция с внешними сервисами

Все HTTP-клиенты для внешних микросервисов находятся в `internal/integrations/`:

**Структура клиента:**
- `client.go` - HTTP клиент с методами API
- `contract.go` - интерфейс Logger
- `errors.go` - sentinel errors (ErrPriceServiceUnavailable, ErrInvalidResponse)
- `models.go` - DTO для запросов/ответов

**Пример клиента (PriceService):**
```go
// internal/integrations/priceservice/client.go
type Client struct {
    baseURL    string
    httpClient *http.Client
    logger     Logger
}

func NewClient(baseURL string, timeout time.Duration, logger Logger) *Client {
    return &Client{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: timeout,
        },
        logger: logger,
    }
}

func (c *Client) GetPrice(ctx context.Context, companyID, serviceID, userID int64) (*PriceResponse, error) {
    url := fmt.Sprintf("%s/api/v1/prices?company_id=%d&service_id=%d&user_id=%d",
        c.baseURL, companyID, serviceID, userID)

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("%w: failed to create request: %v", ErrInvalidRequest, err)
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("%w: %v", ErrPriceServiceUnavailable, err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("%w: status code %d", ErrPriceServiceUnavailable, resp.StatusCode)
    }

    var priceResp PriceResponse
    if err := json.NewDecoder(resp.Body).Decode(&priceResp); err != nil {
        return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
    }

    return &priceResp, nil
}
```

**Инициализация в main.go:**
```go
// Опциональная инициализация клиента
var priceClient *priceservice.Client
if cfg.PriceService.Enabled {
    priceClient = priceservice.NewClient(
        cfg.PriceService.BaseURL,
        time.Duration(cfg.PriceService.TimeoutSeconds)*time.Second,
        log,
    )
}

// Передача в сервисный слой
serviceSvc := services.NewService(serviceRepo, companyRepo, priceClient, log)
```

**Graceful degradation:**
- Если внешний сервис недоступен - логируй WARNING, но не прерывай запрос
- Возвращай данные без обогащения (например, услугу без цены)
- Используй context с таймаутом для предотвращения зависания

## Схема базы данных

PostgreSQL база с ключевыми таблицами:

### companies
```sql
CREATE TABLE companies (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    tags TEXT[],
    logo_url TEXT,
    manager_ids BIGINT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### addresses
```sql
CREATE TABLE addresses (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT REFERENCES companies(id) ON DELETE CASCADE,
    city VARCHAR(255),
    street VARCHAR(255),
    building VARCHAR(50),
    latitude NUMERIC(10, 7),
    longitude NUMERIC(10, 7),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### working_hours
```sql
CREATE TABLE working_hours (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT UNIQUE REFERENCES companies(id) ON DELETE CASCADE,

    -- Каждый день: is_open (boolean), open_time (TIME), close_time (TIME)
    monday_is_open BOOLEAN NOT NULL DEFAULT false,
    monday_open_time TIME,
    monday_close_time TIME,

    tuesday_is_open BOOLEAN NOT NULL DEFAULT false,
    tuesday_open_time TIME,
    tuesday_close_time TIME,

    -- ... аналогично для остальных дней недели

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Важно:** TIME поля в PostgreSQL читаются как time.Time в Go, но должны сериализоваться в JSON как строки "HH:MM". Для этого используется кастомный тип `domain.TimeString` с реализацией `sql.Scanner` и `driver.Valuer`.

### services
```sql
CREATE TABLE services (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    duration_minutes INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### service_addresses
```sql
CREATE TABLE service_addresses (
    service_id BIGINT REFERENCES services(id) ON DELETE CASCADE,
    address_id BIGINT REFERENCES addresses(id) ON DELETE CASCADE,
    PRIMARY KEY (service_id, address_id)
);
```

**Ключевые моменты:**
- Все ID используют `BIGINT` (не UUID)
- Массивы хранятся как PostgreSQL массивы: `TEXT[]`, `BIGINT[]`
- TIME поля читаются через кастомный тип `domain.TimeString` для правильной JSON сериализации
- Работа с массивами через `pq.Array()`:
  ```go
  // Вставка
  pq.Array(tags)           // []string
  pq.Array(managerIDs)     // []int64

  // Сканирование
  var tags pq.StringArray
  var ids pq.Int64Array
  err := row.Scan(&tags, &ids)
  ```

## API дизайн

Спецификация API: `schemas/schema.yaml` (OpenAPI 3.0)

### Endpoints

**Companies (Компании):**
- `POST /api/v1/companies` - Создание (Protected, только superuser)
- `GET /api/v1/companies` - Список с фильтрами (Public)
- `GET /api/v1/companies/{id}` - Получение по ID (Public)
- `PUT /api/v1/companies/{id}` - Обновление (Protected, superuser или manager)
- `DELETE /api/v1/companies/{id}` - Удаление (Protected, только superuser)

**Services (Услуги):**
- `POST /api/v1/companies/{company_id}/services` - Создание (Protected, superuser или manager компании)
- `GET /api/v1/companies/{company_id}/services` - Список (Public, опциональный `X-User-ID` для получения цен)
- `GET /api/v1/companies/{company_id}/services/{service_id}` - Получение по ID (Public, опциональный `X-User-ID` для получения цен)
- `PUT /api/v1/companies/{company_id}/services/{service_id}` - Обновление (Protected, superuser или manager)
- `DELETE /api/v1/companies/{company_id}/services/{service_id}` - Удаление (Protected, superuser или manager)

**Интеграция с ценами (PriceService):**
GET endpoints для услуг поддерживают опциональный заголовок `X-User-ID`. Если он передан:
- SellerService отправляет запрос в SMC-PriceService для получения цен
- Ответ обогащается полями: `price`, `currency`, `pricing_type`, `vehicle_class`, `applied_multiplier`
- Если PriceService недоступен или цены нет - поля остаются `null`, но ответ всё равно успешный

### Аутентификация (MVP - Упрощённая версия)

Защищённые endpoints требуют два заголовка:
```
X-User-ID: <user_id>
X-User-Role: <superuser|user>
```

**⚠️ Важно**: Это временное решение для MVP. В продакшене планируется:
- Отдельный SMC-AuthService для генерации JWT токенов
- Валидация Telegram InitData
- Refresh token механизм
- Полноценная JWT аутентификация

**Middleware `auth.go`:**
```go
func UserIDAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userIDStr := r.Header.Get("X-User-ID")
        userRole := r.Header.Get("X-User-Role")

        userID, err := strconv.ParseInt(userIDStr, 10, 64)
        if err != nil || userID <= 0 {
            // ... error handling
            return
        }

        ctx := context.WithValue(r.Context(), userIDKey, userID)
        ctx = context.WithValue(ctx, userRoleKey, userRole)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Извлечение из контекста
userID, ok := middleware.GetUserID(r.Context())
userRole, ok := middleware.GetUserRole(r.Context())
```

### Правила авторизации

**superuser** (`constants.RoleSuperuser`):
- Полный доступ ко всем операциям
- Может создавать/удалять компании
- Может изменять любые компании и услуги

**user** (`constants.RoleUser`):
- Может изменять только компании, где он указан в массиве `manager_ids`
- Может создавать/изменять/удалять услуги только для своих компаний
- Проверка доступа: `companyRepo.IsManager(ctx, companyID, userID)`

**Примеры проверки:**
```go
// В сервисном слое
func (s *Service) Update(ctx context.Context, userID int64, userRole string, id int64, req *models.UpdateCompanyRequest) (*models.CompanyResponse, error) {
    if userRole == constants.RoleSuperuser {
        // Superuser может всё
        // ... обновление
    }

    // User может только свои компании
    isManager, err := s.companyRepo.IsManager(ctx, id, userID)
    if err != nil {
        return nil, fmt.Errorf("%w: IsManager check failed: %v", ErrInternal, err)
    }
    if !isManager {
        return nil, ErrAccessDenied
    }
    // ... обновление
}
```

### Паттерны ответов

- Все endpoints возвращают ОДИНАКОВУЮ схему (нет отдельных "публичных" и "полных" данных)
- List endpoints поддерживают опциональную пагинацию через query параметры `page` и `limit`
  - Нет пагинации = вернуть все результаты
- Фильтр по тегам: через запятую (например, `?tags=#мойка,#москва`)
- Стандартные коды ответов:
  - 200 OK - успешное получение
  - 201 Created - успешное создание
  - 400 Bad Request - невалидный request body
  - 401 Unauthorized - отсутствуют заголовки аутентификации
  - 403 Forbidden - недостаточно прав
  - 404 Not Found - ресурс не найден
  - 500 Internal Server Error - внутренняя ошибка сервера

## Соглашения по коду

### Обработка ошибок (трёхуровневая система)

#### 1. Repository Layer

**errors.go:**
```go
var (
    ErrCompanyNotFound = errors.New("repository: company not found")
    ErrBuildQuery = errors.New("repository: failed to build SQL query")
    ErrExecQuery = errors.New("repository: failed to execute SQL query")
    ErrScanRow = errors.New("repository: failed to scan row")
    ErrTransaction = errors.New("repository: transaction error")
)
```

**repository.go:**
```go
func (r *Repository) GetByID(ctx context.Context, id int64) (*domain.Company, error) {
    query, args, err := psqlbuilder.Select("id", "name").
        From("companies").
        Where(squirrel.Eq{"id": id}).
        ToSql()
    if err != nil {
        return nil, fmt.Errorf("%w: GetByID - build query: %v", ErrBuildQuery, err)
    }

    var company domain.Company
    err = r.db.QueryRowContext(ctx, query, args...).Scan(&company.ID, &company.Name)
    if err == sql.ErrNoRows {
        return nil, ErrCompanyNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("%w: GetByID - scan company: %v", ErrScanRow, err)
    }

    return &company, nil
}
```

#### 2. Service Layer

**errors.go:**
```go
var (
    ErrCompanyNotFound = errors.New("company not found")
    ErrAccessDenied = errors.New("access denied: user is not a manager of this company")
    ErrOnlySuperuser = errors.New("access denied: only superuser can create companies")
    ErrInvalidInput = errors.New("invalid input data")
    ErrInternal = errors.New("service: internal error")
)
```

**service.go:**
```go
import (
    companyRepo "github.com/m04kA/SMC-SellerService/internal/infra/storage/company"
)

func (s *Service) GetByID(ctx context.Context, id int64) (*models.CompanyResponse, error) {
    company, err := s.companyRepo.GetByID(ctx, id)
    if err != nil {
        // Проверка через errors.Is()
        if errors.Is(err, companyRepo.ErrCompanyNotFound) {
            return nil, ErrCompanyNotFound
        }
        // Оборачивание неизвестных ошибок
        return nil, fmt.Errorf("%w: GetByID - repository error: %v", ErrInternal, err)
    }

    return models.FromDomainCompany(company), nil
}
```

#### 3. Handler Layer

**handler.go:**
```go
const (
    msgInvalidRequestBody = "invalid request body"
    msgForbidden          = "access denied"
    msgNotFound           = "company not found"
)

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
    company, err := h.service.GetByID(r.Context(), id)
    if err != nil {
        // Проверка через errors.Is()
        if errors.Is(err, companies.ErrCompanyNotFound) {
            h.logger.Warn("GET /companies/%d - Not found", id)
            handlers.RespondNotFound(w, msgNotFound)
            return
        }
        if errors.Is(err, companies.ErrAccessDenied) {
            h.logger.Warn("GET /companies/%d - Access denied: user_id=%d", id, userID)
            handlers.RespondForbidden(w, msgForbidden)
            return
        }
        h.logger.Error("GET /companies/%d - Internal error: %v", id, err)
        handlers.RespondInternalError(w)
        return
    }

    h.logger.Info("GET /companies/%d - Success", id)
    handlers.RespondJSON(w, http.StatusOK, company)
}
```

### Пути импортов

Всегда используй полный путь модуля:
```go
import (
    "github.com/m04kA/SMC-SellerService/internal/domain"
    "github.com/m04kA/SMC-SellerService/internal/service/companies/models"
    "github.com/m04kA/SMC-SellerService/pkg/psqlbuilder"
    "github.com/m04kA/SMC-SellerService/pkg/logger"

    // Импорт репозиториев с алиасами
    companyRepo "github.com/m04kA/SMC-SellerService/internal/infra/storage/company"
    serviceRepo "github.com/m04kA/SMC-SellerService/internal/infra/storage/service"

    "github.com/Masterminds/squirrel"
    "github.com/lib/pq"
    "github.com/gorilla/mux"
)
```

### Dependency Injection в main.go

```go
func main() {
    // 1. Config
    cfg, err := config.Load("config.toml")
    if err != nil {
        panic(err)
    }

    // 2. Logger
    log, err := logger.New("./logs/app.log")
    if err != nil {
        panic(err)
    }
    defer log.Close()

    // 3. Database
    db, err := sql.Open("postgres", cfg.Database.DSN())
    if err != nil {
        panic(err)
    }
    defer db.Close()

    // 4. Repositories
    companyRepo := company.NewRepository(db)
    serviceRepo := service.NewRepository(db)

    // 5. Services
    companySvc := companies.NewService(companyRepo)
    serviceSvc := services.NewService(serviceRepo, companyRepo)

    // 6. Handlers
    createCompanyHandler := create_company.NewHandler(companySvc, log)
    getCompanyHandler := get_company.NewHandler(companySvc, log)
    // ... остальные handlers

    // 7. Routing
    router := mux.NewRouter()

    // Public routes
    public := router.PathPrefix("/api/v1").Subrouter()
    public.HandleFunc("/companies", listCompaniesHandler.Handle).Methods(http.MethodGet)
    public.HandleFunc("/companies/{id}", getCompanyHandler.Handle).Methods(http.MethodGet)
    // ...

    // Protected routes
    protected := router.PathPrefix("/api/v1").Subrouter()
    protected.Use(middleware.UserIDAuth)
    protected.HandleFunc("/companies", createCompanyHandler.Handle).Methods(http.MethodPost)
    protected.HandleFunc("/companies/{id}", updateCompanyHandler.Handle).Methods(http.MethodPut)
    // ...

    // 8. Server
    srv := &http.Server{
        Addr:    fmt.Sprintf(":%d", cfg.Server.HTTPPort),
        Handler: router,
    }

    // 9. Graceful shutdown
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Error("Server error: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        log.Error("Server shutdown error: %v", err)
    }
}
```

## Configuration

### config.toml

```toml
[logs]
level = "info"

[server]
http_port = 8081

[database]
host = "localhost"
port = 5436
user = "postgres"
password = "postgres"
dbname = "smc_sellerservice"
sslmode = "disable"

[price_service]
base_url = "http://localhost:8082"
timeout_seconds = 5
enabled = true
```

**Секции конфигурации:**
- `[logs]` - настройки логирования
- `[server]` - HTTP порт сервиса
- `[database]` - подключение к PostgreSQL
- `[price_service]` - интеграция с SMC-PriceService (опционально, если `enabled = false` - цены не запрашиваются)

### Docker Compose

**docker-compose.yml** (корневая директория):
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: smc-sellerservice-db
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: smc_sellerservice
    ports:
      - "5436:5432"
    volumes:
      - ./docker/postgres/data:/var/lib/postgresql/data
    networks:
      - smc-sellerservice-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  migrate:
    image: migrate/migrate
    container_name: smc-sellerservice-migrate
    depends_on:
      postgres:
        condition: service_healthy
    volumes:
      - ./migrations:/migrations
    networks:
      - smc-sellerservice-network
    command: [
      "-path", "/migrations",
      "-database", "postgres://postgres:postgres@postgres:5432/smc_sellerservice?sslmode=disable",
      "up"
    ]
    restart: on-failure

networks:
  smc-sellerservice-network:
    driver: bridge
```

**Ключевые моменты:**
- PostgreSQL порт **5436** на хосте (5432 в контейнере)
- Автоматическое применение миграций через контейнер `migrate`
- Healthcheck для postgres перед запуском миграций
- Volumes для персистентности данных

## Миграции базы данных

Миграции находятся в директории `migrations/` с паттерном именования:
- `000001_init_schema.up.sql` - Применение миграции
- `000001_init_schema.down.sql` - Откат миграции

**Применение миграций:**
```bash
# Автоматически через docker-compose
docker-compose up -d

# Вручную
docker-compose run --rm migrate -path /migrations -database "postgres://postgres:postgres@postgres:5432/smc_sellerservice?sslmode=disable" up

# Откат
docker-compose run --rm migrate -path /migrations -database "postgres://postgres:postgres@postgres:5432/smc_sellerservice?sslmode=disable" down
```

## Интеграция с микросервисами

Этот сервис является частью микросервисной архитектуры SMC (Slot My Car):

### SMC-UserService (порт 8080)

**Общие паттерны:**
- Clean Architecture с тем же разделением слоёв
- Аутентификация через заголовки `X-User-ID` и `X-User-Role`
- Handler per Endpoint паттерн
- Injectable Logger с интерфейсами в contract.go
- Трёхуровневая система обработки ошибок
- Gorilla Mux для роутинга
- psqlbuilder для построения SQL запросов
- PostgreSQL с нестандартным портом (UserService: 5435, SellerService: 5436)
- golang-migrate для миграций
- Docker Compose для контейнеризации

**Роли в UserService:**
- `client` (role_id=1) - клиент, может видеть только свои данные
- `manager` (role_id=2) - менеджер компании, может управлять компанией
- `superuser` (role_id=3) - администратор, полный доступ

**Роли в SellerService:**
- `user` - менеджер компании (соответствует `manager` из UserService)
- `superuser` - администратор (соответствует `superuser` из UserService)

**Взаимодействие:**
- SellerService получает `userID` и `userRole` через заголовки от клиента
- В будущем планируется SMC-AuthService для JWT токенов
- UserService предоставляет endpoint `GET /internal/users/{tg_user_id}` для получения информации о пользователе

### SMC-PriceService (порт 8082) ✅

**Назначение:**
- Управление ценами на услуги в зависимости от класса автомобиля пользователя
- Предоставляет endpoint для получения цены: `GET /api/v1/prices?company_id={company_id}&service_id={service_id}&user_id={user_id}`

**Интеграция с SellerService:**
- Клиент находится в `internal/integrations/priceclient/`
- Конфигурируется через `config.toml` в секции `[price_service]`
- GET endpoints для услуг (`/companies/{id}/services` и `/companies/{id}/services/{id}`) автоматически обогащаются ценами при передаче заголовка `X-User-ID`
- При недоступности PriceService ошибка логируется как WARNING, но запрос успешно возвращает данные услуги без цен
- Таймаут HTTP клиента: 5 секунд (настраивается)

**Структура ответа от PriceService:**
```json
{
  "price": 1500.00,
  "currency": "RUB",
  "pricing_type": "fixed",
  "vehicle_class": "sedan",
  "applied_multiplier": 1.2
}
```

**Пример кода интеграции:**
```go
// internal/service/services/service.go
func (s *Service) GetByID(ctx context.Context, companyID, serviceID int64, userID *int64) (*models.ServiceResponse, error) {
    service, err := s.serviceRepo.GetByID(ctx, companyID, serviceID)
    if err != nil {
        // ... error handling
    }

    response := models.FromDomainService(service)

    // Обогащение ценой, если передан userID
    if userID != nil && s.priceClient != nil {
        priceResp, err := s.priceClient.GetPrice(ctx, companyID, serviceID, *userID)
        if err != nil {
            s.logger.Warn("Failed to get price from PriceService: %v", err)
        } else if priceResp != nil {
            response.EnrichWithPrice(
                priceResp.Price,
                priceResp.Currency,
                priceResp.PricingType,
                priceResp.VehicleClass,
                priceResp.AppliedMultiplier,
            )
        }
    }

    return response, nil
}
```

### Будущая архитектура (с аутентификацией)

```
Telegram Bot
    ↓
SMC-AuthService (JWT генерация, валидация InitData)
    ↓
API Gateway (валидация JWT, маршрутизация)
    ↓
├── SMC-UserService (пользователи, автомобили)
├── SMC-SellerService (компании, услуги)
└── SMC-PriceService (цены)
```

## Best Practices

### DO ✅

1. **Используй sentinel errors с errors.Is():**
   ```go
   if errors.Is(err, companyRepo.ErrCompanyNotFound) {
       return nil, ErrCompanyNotFound
   }
   ```

2. **Оборачивай ошибки с контекстом:**
   ```go
   return nil, fmt.Errorf("%w: GetByID - scan company: %v", ErrScanRow, err)
   ```

3. **Используй константы для сообщений в handlers:**
   ```go
   const (
       msgInvalidRequestBody = "invalid request body"
       msgForbidden          = "access denied"
   )
   ```

4. **Логируй с контекстом:**
   ```go
   h.logger.Info("POST /companies - Created: company_id=%d, user_id=%d", company.ID, userID)
   h.logger.Error("POST /companies - Failed: user_id=%d, error=%v", userID, err)
   ```

5. **Используй squirrel.Eq для WHERE условий:**
   ```go
   Where(squirrel.Eq{"id": id, "company_id": companyID})
   ```

6. **Явный rollback при ошибках в транзакциях:**
   ```go
   if err != nil {
       tx.Rollback()
       return nil, err
   }
   ```

7. **Извлекай только нужные колонки:**
   ```go
   psqlbuilder.Select("id", "name", "description") // не "SELECT *"
   ```

8. **Используй алиасы для импортов репозиториев:**
   ```go
   import companyRepo "github.com/m04kA/SMC-SellerService/internal/infra/storage/company"
   ```

9. **Graceful degradation для внешних сервисов:**
   ```go
   // ✅ Правильно - логируй WARNING, но не прерывай запрос
   if userID != nil && s.priceClient != nil {
       priceResp, err := s.priceClient.GetPrice(ctx, companyID, serviceID, *userID)
       if err != nil {
           s.logger.Warn("Failed to get price from PriceService: %v", err)
           // Продолжаем без цены
       } else if priceResp != nil {
           response.EnrichWithPrice(priceResp.Price, priceResp.Currency, ...)
       }
   }
   ```

10. **Используй context с таймаутом в HTTP клиентах:**
    ```go
    client := &http.Client{
        Timeout: 5 * time.Second,
    }
    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    ```

### DON'T ❌

1. **Не используй глобальный logger:**
   ```go
   // ❌ Неправильно
   logger.Info("message")

   // ✅ Правильно
   log, err := logger.New("./logs/app.log")
   log.Info("message")
   ```

2. **Не используй string comparison для ошибок:**
   ```go
   // ❌ Неправильно
   if strings.Contains(err.Error(), "not found") {
       // ...
   }

   // ✅ Правильно
   if errors.Is(err, companyRepo.ErrCompanyNotFound) {
       // ...
   }
   ```

3. **Не возвращай текст ошибки в HTTP ответах:**
   ```go
   // ❌ Неправильно
   handlers.RespondNotFound(w, err.Error())

   // ✅ Правильно
   handlers.RespondNotFound(w, msgNotFound)
   ```

4. **Не используй defer для rollback:**
   ```go
   // ❌ Неправильно
   defer tx.Rollback()

   // ✅ Правильно
   if err != nil {
       tx.Rollback()
       return nil, err
   }
   ```

5. **Не используй плейсхолдеры напрямую:**
   ```go
   // ❌ Неправильно
   Where("id = ?", id)

   // ✅ Правильно
   Where(squirrel.Eq{"id": id})
   ```

6. **Не добавляй бизнес-логику в domain модели:**
   ```go
   // ❌ Неправильно (в domain/company.go)
   func (c *Company) CanBeModifiedBy(userID int64) bool { ... }

   // ✅ Правильно (в service/companies/service.go)
   func (s *Service) checkAccess(userID int64, companyID int64) error { ... }
   ```

7. **Не используй SELECT * в запросах:**
   ```go
   // ❌ Неправильно
   psqlbuilder.Select("*").From("companies")

   // ✅ Правильно
   psqlbuilder.Select("id", "name", "description").From("companies")
   ```

8. **Не прерывай запрос при недоступности опциональных внешних сервисов:**
   ```go
   // ❌ Неправильно - прерываем запрос
   priceResp, err := s.priceClient.GetPrice(ctx, companyID, serviceID, *userID)
   if err != nil {
       return nil, fmt.Errorf("failed to get price: %v", err)
   }

   // ✅ Правильно - возвращаем данные без цены
   priceResp, err := s.priceClient.GetPrice(ctx, companyID, serviceID, *userID)
   if err != nil {
       s.logger.Warn("Failed to get price: %v", err)
       // Продолжаем без цены
   }
   ```

## Troubleshooting

### Порт уже занят

```bash
# Проверка занятых портов
lsof -i :8081  # SellerService
lsof -i :5436  # PostgreSQL

# Остановка контейнеров
docker-compose down
```

### Миграции не применяются

```bash
# Проверка логов
docker-compose logs migrate

# Ручное применение
docker-compose up -d postgres
sleep 5
docker-compose up migrate
```

### Ошибки подключения к БД

```bash
# Проверка конфигурации
cat config.toml

# При запуске в Docker используется host="postgres" port=5432
# При локальном запуске используется host="localhost" port=5436
```

### Логи не пишутся

```bash
# Создание директории для логов
mkdir -p logs

# Проверка прав доступа
ls -la logs/
chmod 755 logs/
```

## Информация о модуле

- **Имя модуля**: `github.com/m04kA/SMC-SellerService`
- **Версия Go**: 1.24.2
- **Ключевые зависимости**:
  - `github.com/Masterminds/squirrel` v1.5.4 - SQL query builder
  - `github.com/lib/pq` v1.10.9 - PostgreSQL драйвер
  - `github.com/gorilla/mux` v1.8.1 - HTTP router

## Тестирование API

### Документация тестов
- **Curl команды**: `test_data/API_TEST_COMMANDS.md` - все тесты с описанием
- **Bruno коллекция**: `/Users/yapanarin/GolandProjects/SMC-Bruno/SMC/SellerService/` - 16 готовых запросов

### Тестовые данные
Все тестовые JSON файлы находятся в `test_data/`:
- `create_company.json` - создание компании
- `create_company_2.json` - создание второй компании (для DELETE)
- `update_company.json` - обновление компании
- `create_service.json` - создание услуги
- `update_service.json` - обновление услуги

### Проверенные сценарии
✅ Все CRUD операции для Companies
✅ Все CRUD операции для Services
✅ Авторизация (superuser vs user)
✅ Фильтрация по тегам и городу
✅ Edge cases (404, 403, невалидные данные)
✅ Формат данных соответствует OpenAPI schema
✅ WorkingHours возвращаются в формате "HH:MM"

## Дополнительные ресурсы

- **OpenAPI спецификация**: `schemas/schema.yaml`
- **Миграции**: `migrations/`
- **Docker Compose**: `docker-compose.yml` (корневая директория)
- **Тестовые данные**: `test_data/`
- **Bruno коллекция**: `/Users/yapanarin/GolandProjects/SMC-Bruno/SMC/SellerService/`
- **Референсный проект**: SMC-UserService (`/Users/yapanarin/GolandProjects/SMC-UserService`)
