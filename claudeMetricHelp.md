# План внедрения централизованного мониторинга для SMC

## Цель
Создать централизованную систему мониторинга для всех микросервисов SMC (UserService, SellerService, PriceService, AuthService) с разделением метрик по дашбордам.

## Архитектура

```
┌─────────────────────────────────────────────────────────┐
│              SMC-Monitoring (порт 3000)                  │
│                                                           │
│  ┌─────────────┐         ┌──────────────┐               │
│  │ Prometheus  │◄────────┤   Grafana    │               │
│  │  :9090      │         │    :3000     │               │
│  │             │         │              │               │
│  │ - scrapes   │         │ Dashboards:  │               │
│  │   metrics   │         │ - Overview   │               │
│  │ - stores    │         │ - UserSvc    │               │
│  │   timeseries│         │ - SellerSvc  │               │
│  │             │         │ - PriceSvc   │               │
│  └──────▲──────┘         │ - AuthSvc    │               │
│         │                │ - Postgres   │               │
│         │                └──────────────┘               │
└─────────┼──────────────────────────────────────────────┘
          │
          │ scrape /metrics every 15s
          │
    ┌─────┴─────┬─────────┬─────────┬─────────┐
    │           │         │         │         │
┌───▼────┐  ┌──▼────┐ ┌──▼────┐ ┌──▼────┐ ┌──▼────────┐
│  User  │  │Seller │ │ Price │ │ Auth  │ │ Postgres  │
│Service │  │Service│ │Service│ │Service│ │ Exporter  │
│ :8080  │  │ :8081 │ │ :8082 │ │ :8083 │ │  :9187    │
└────────┘  └───────┘ └───────┘ └───────┘ └───────────┘
 /metrics    /metrics  /metrics  /metrics   /metrics
```

## Этапы реализации

### Этап 1: Подготовка SMC-SellerService ✅
**Что делаем в этом репозитории:**

1. **Создать pkg/metrics/metrics.go**
   - Prometheus клиент для сбора метрик
   - HTTP метрики: запросы, ошибки, latency
   - Database метрики: запросы, ошибки, latency, pool stats

2. **Создать middleware для HTTP метрик**
   - `internal/api/middleware/metrics.go`
   - Сбор метрик для каждого endpoint
   - Разделение по методам (GET, POST, PUT, DELETE)
   - Разделение по статус кодам (2xx, 4xx, 5xx)

3. **Добавить wrapper для Database метрик**
   - `pkg/dbmetrics/dbmetrics.go`
   - Обёртка над sql.DB для сбора метрик
   - Метрики: query duration, errors, active connections, idle connections

4. **Добавить /metrics endpoint**
   - В cmd/main.go добавить роут GET /metrics
   - Expose Prometheus метрики

5. **Обновить config.toml**
   - Добавить секцию [metrics]
   - enabled = true/false
   - path = "/metrics"

### Этап 2: Создание SMC-Monitoring (новый репозиторий)
**Что создаём:**

```
SMC-Monitoring/
├── README.md
├── docker-compose.yml
├── .env.example
├── prometheus/
│   ├── prometheus.yml
│   └── alerts/
│       └── rules.yml
├── grafana/
│   ├── provisioning/
│   │   ├── datasources/
│   │   │   └── prometheus.yml
│   │   └── dashboards/
│   │       └── dashboards.yml
│   └── dashboards/
│       ├── 00-overview.json
│       ├── 01-userservice.json
│       ├── 02-sellerservice.json
│       ├── 03-priceservice.json
│       ├── 04-authservice.json
│       └── 05-postgres.json
└── postgres-exporter/
    └── queries.yaml
```

**docker-compose.yml будет содержать:**
- Prometheus (порт 9090)
- Grafana (порт 3000)
- postgres-exporter (порт 9187) - для метрик PostgreSQL

### Этап 3: Настройка Prometheus
**prometheus.yml будет содержать scrape configs для:**
- SMC-UserService (localhost:8080/metrics)
- SMC-SellerService (localhost:8081/metrics)
- SMC-PriceService (localhost:8082/metrics)
- SMC-AuthService (localhost:8083/metrics)
- postgres-exporter для UserService DB
- postgres-exporter для SellerService DB

**Алерты (alerts/rules.yml):**
- High error rate (5xx > 5%)
- High latency (p95 > 1s)
- Database connection pool exhausted
- Service down

### Этап 4: Создание Grafana Dashboards

#### Dashboard 1: Overview (00-overview.json)
**Панели:**
- Total Requests (все сервисы)
- Error Rate (все сервисы)
- Average Response Time (все сервисы)
- Services Status (UP/DOWN)
- Top 10 Slowest Endpoints

#### Dashboard 2: SellerService (02-sellerservice.json)
**Структура по вашему требованию:**

**Row 1: 2XX Responses**
- Panel 1: RPM (Requests Per Minute) для 2XX
- Panel 2: Response Time (p50, p95, p99) для 2XX

**Row 2: 4XX Responses**
- Panel 1: RPM для 4XX
- Panel 2: Response Time для 4XX

**Row 3: 5XX Responses**
- Panel 1: RPM для 5XX
- Panel 2: Response Time для 5XX

**Row 4: Database Metrics**
- Panel 1: Database Queries Per Second
- Panel 2: Database Query Duration (p50, p95, p99)
- Panel 3: Database Errors Count
- Panel 4: Database Connection Pool (active/idle/max)

**Row 5: Endpoints Breakdown**
- Table: Top endpoints by response time
- Table: Top endpoints by error rate

#### Dashboard 3: Postgres (05-postgres.json)
**Панели:**
- Active Connections
- Queries Per Second
- Transaction Rate
- Cache Hit Ratio
- Dead Tuples
- Table Size
- Index Usage

### Этап 5: Метрики которые собираем

#### HTTP Metrics (для каждого сервиса)
```
# Счётчик запросов
http_requests_total{service="sellerservice", method="GET", endpoint="/api/v1/companies", status_code="200"}

# Гистограмма времени ответа
http_request_duration_seconds{service="sellerservice", method="GET", endpoint="/api/v1/companies", status_code="200"}

# Счётчик ошибок
http_errors_total{service="sellerservice", method="POST", endpoint="/api/v1/companies", status_code="500", error_type="internal"}
```

#### Database Metrics
```
# Гистограмма времени запросов
db_query_duration_seconds{service="sellerservice", operation="select", table="companies"}

# Счётчик запросов
db_queries_total{service="sellerservice", operation="select", table="companies", status="success"}

# Счётчик ошибок
db_errors_total{service="sellerservice", operation="insert", table="companies", error_type="constraint_violation"}

# Gauge для connection pool
db_connections_active{service="sellerservice"}
db_connections_idle{service="sellerservice"}
db_connections_max{service="sellerservice"}
```

## Текущий статус

### ✅ Этап 1: Подготовка SMC-SellerService (ЗАВЕРШЁН!)
- [x] Создать pkg/metrics/metrics.go
- [x] Создать middleware для HTTP метрик
- [x] Добавить wrapper для Database метрик
- [x] Добавить /metrics endpoint в main.go
- [x] Обновить config.toml
- [x] Обновить internal/config/config.go для поддержки MetricsConfig
- [x] Обновить .env и .env.example
- [x] Добавить prometheus_client в go.mod (v1.23.2)
- [x] Успешная сборка приложения (bin/sellerservice)

**Результаты Этапа 1:**

**Файловая структура:**
```
pkg/
├── metrics/
│   └── metrics.go              # ✅ Prometheus метрики (HTTP + DB)
├── dbmetrics/
│   └── dbmetrics.go            # ✅ Database wrapper с метриками
└── logger/
    └── logger.go

internal/
├── api/
│   └── middleware/
│       ├── auth.go
│       └── metrics.go          # ✅ HTTP metrics middleware
├── config/
│   └── config.go               # ✅ Добавлен MetricsConfig
└── ...

cmd/
└── main.go                      # ✅ Интегрированы метрики

config.toml                      # ✅ Секция [metrics]
.env                             # ✅ METRICS_ENABLED=true
.env.example                     # ✅ Документация метрик
```

**Endpoints:**
- ✅ `GET /metrics` - Prometheus метрики (публичный, без auth)
- ✅ Все API endpoints оборачиваются в metrics middleware

**Метрики которые собираются:**

**HTTP:**
- `http_requests_total{service="sellerservice", method="GET", endpoint="/api/v1/companies", status_code="200"}`
- `http_request_duration_seconds{service="sellerservice", method="GET", endpoint="/api/v1/companies", status_code="200"}`
- `http_errors_total{service="sellerservice", method="POST", status_code="500", error_type="internal_error"}`

**Database:**
- `db_queries_total{service="sellerservice", operation="select", table="companies", status="success"}`
- `db_query_duration_seconds{service="sellerservice", operation="select", table="companies"}`
- `db_errors_total{service="sellerservice", operation="insert", table="companies", error_type="duplicate_key"}`
- `db_connections_active{service="sellerservice"}` (обновляется каждые 15 секунд)
- `db_connections_idle{service="sellerservice"}`
- `db_connections_max{service="sellerservice"}`

**Конфигурация (config.toml):**
```toml
[metrics]
enabled = true                 # Включить сбор метрик
path = "/metrics"              # Путь для Prometheus
service_name = "sellerservice" # Имя сервиса в метриках
```

**Использование:**
```bash
# Запуск приложения
go run cmd/main.go

# Проверка метрик
curl http://localhost:8081/metrics

# Выключить метрики
METRICS_ENABLED=false go run cmd/main.go
```

### ⏳ Этап 2: Создание SMC-Monitoring (СЛЕДУЮЩИЙ)
- [ ] Создать новый репозиторий `/Users/yapanarin/GolandProjects/SMC-Monitoring`
- [ ] Настроить docker-compose.yml (Prometheus + Grafana + postgres-exporter)
- [ ] Создать prometheus/prometheus.yml с scrape configs
- [ ] Настроить Grafana provisioning (datasources + dashboards)
- [ ] Создать базовую структуру директорий

### ⏳ Этап 3: Grafana Dashboards
- [ ] Dashboard: Overview (все сервисы)
- [ ] Dashboard: SellerService (с 2XX/4XX/5XX панелями + DB метрики)
- [ ] Dashboard: Postgres (для всех баз данных)
- [ ] Dashboard: UserService
- [ ] Dashboard: PriceService
- [ ] Dashboard: AuthService

### ⏳ Этап 4: Интеграция
- [ ] Подключить SMC-SellerService к Prometheus
- [ ] Подключить остальные сервисы (UserService, PriceService, AuthService)
- [ ] Настроить алерты (error rate, latency, service down)
- [ ] Создать README.md с документацией
- [ ] Настроить retention policy для метрик

## Преимущества такого подхода

1. **Централизация**: Все метрики в одном месте
2. **Масштабируемость**: Легко добавить новый сервис
3. **Корреляция**: Видно как сервисы влияют друг на друга
4. **Разделение**: Каждый сервис имеет свой дашборд
5. **Экономия ресурсов**: Один Prometheus + одна Grafana для всех
6. **Удобство**: Не нужно переключаться между разными Grafana

## Следующие шаги

1. ✅ **Этап 1 ЗАВЕРШЁН** - Метрики добавлены в SMC-SellerService
2. 🚀 **Этап 2 СЛЕДУЮЩИЙ** - Создание SMC-Monitoring репозитория
3. ⏳ Повторить Этап 1 для остальных сервисов (UserService, PriceService, AuthService)
4. ⏳ Подключить все сервисы к централизованному мониторингу

## Вопросы для обсуждения

- [ ] Где будет работать SMC-Monitoring? (localhost, отдельный сервер?)
- [ ] Нужны ли алерты в Telegram/Email?
- [ ] Какой retention period для метрик? (15 дней по умолчанию)
- [ ] Нужна ли аутентификация для Grafana?

---

## Проверка работоспособности метрик

**Запуск приложения:**
```bash
# С метриками (по умолчанию)
go run cmd/main.go

# Проверка /metrics endpoint
curl http://localhost:8081/metrics
```

**Ожидаемый вывод:**
```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{endpoint="/api/v1/companies",method="GET",service="sellerservice",status_code="200"} 5

# HELP http_request_duration_seconds HTTP request duration in seconds
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{endpoint="/api/v1/companies",method="GET",service="sellerservice",status_code="200",le="0.005"} 3
...

# HELP db_queries_total Total number of database queries
# TYPE db_queries_total counter
db_queries_total{operation="select",service="sellerservice",status="success",table="companies"} 10

# HELP db_connections_active Number of active database connections
# TYPE db_connections_active gauge
db_connections_active{service="sellerservice"} 2
...
```

---

**✅ Этап 1: Подготовка SMC-SellerService - ЗАВЕРШЁН!**
**🚀 Готовы к Этапу 2: Создание SMC-Monitoring**
