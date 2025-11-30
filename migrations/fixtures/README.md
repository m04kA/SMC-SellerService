# SellerService - Test Fixtures

Тестовые данные для интеграции с BookingService.

## Загрузка фикстур

```bash
# Рекомендуемый способ (через Makefile)
make fixtures-load

# Альтернативный способ (через docker exec)
docker-compose exec -T postgres psql -U postgres -d smc_sellerservice -f - < migrations/fixtures/001_test_companies.sql
```

**Преимущества `make fixtures-load`:**
- Автоматически запускает PostgreSQL контейнер если он не запущен
- Применяет все файлы фикстур из директории `migrations/fixtures/`
- Использует `ON CONFLICT` для безопасного повторного применения
- Сбрасывает sequences для корректной генерации новых ID

## Содержимое

### 001_test_companies.sql

**3 компании, 4 адреса, 5 услуг**

#### Компания 1: Автомойка Премиум (ID: 1)
- **Менеджер**: 777777777
- **Адреса**:
  - ID 100: Москва, ул. Тверская, 10
  - ID 101: Москва, ул. Ленина, 5
- **Услуги**:
  - ID 1: Комплексная мойка (60 мин) - доступна на обоих адресах
  - ID 2: Экспресс-мойка (30 мин) - только на адресе 100
  - ID 3: Детейлинг (120 мин) - только на адресе 101
- **Рабочие часы**: Пн-Пт 09:00-21:00, Сб 10:00-20:00, Вс выходной

#### Компания 2: СТО Профи (ID: 2)
- **Менеджер**: 888888888
- **Адреса**:
  - ID 200: Москва, ул. Новая, 15
- **Услуги**:
  - ID 10: Замена масла (45 мин)
- **Рабочие часы**: Круглосуточно

#### Компания 3: Детейлинг Центр (ID: 3)
- **Менеджер**: 999999000
- **Адреса**:
  - ID 300: Санкт-Петербург, Невский пр., 1
- **Услуги**:
  - ID 20: Полировка кузова (180 мин)
- **Рабочие часы**: Пн-Вс 08:00-22:00

## Проверка загруженных данных

```sql
-- Подключиться к БД
psql -U postgres -d smc_sellerservice

-- Проверить компании
SELECT id, name, manager_ids FROM companies;

-- Проверить адреса
SELECT id, company_id, city, street, building FROM addresses;

-- Проверить услуги и их адреса
SELECT s.id, s.company_id, s.name, s.average_duration,
       array_agg(sa.address_id) as address_ids
FROM services s
LEFT JOIN service_addresses sa ON s.id = sa.service_id
GROUP BY s.id, s.company_id, s.name, s.average_duration
ORDER BY s.id;

-- Проверить рабочие часы
SELECT company_id,
       monday_is_open, monday_open_time, monday_close_time,
       sunday_is_open
FROM working_hours;
```

## Интеграция с BookingService

Эти данные полностью совместимы с фикстурами BookingService.

### Соответствие данных

| BookingService | SellerService |
|----------------|---------------|
| company_id: 1 | Автомойка Премиум |
| company_id: 2 | СТО Профи |
| company_id: 3 | Детейлинг Центр |
| address_id: 100 | Тверская (компания 1) |
| address_id: 101 | Ленина (компания 1) |
| address_id: 200 | Новая (компания 2) |
| address_id: 300 | Невский пр. (компания 3) |
| service_id: 1 | Комплексная мойка |
| service_id: 2 | Экспресс-мойка |
| service_id: 3 | Детейлинг |
| service_id: 10 | Замена масла |
| service_id: 20 | Полировка кузова |

### Важные моменты

- **Услуга 2** (Экспресс-мойка) доступна ТОЛЬКО на адресе 100
- **Услуга 3** (Детейлинг) доступна ТОЛЬКО на адресе 101
- **Услуга 1** (Комплексная мойка) доступна на обоих адресах

Это используется в BookingService для тестирования проверки `addressIds`.

## Сброс данных

```bash
# Удалить все данные
docker-compose exec -T postgres psql -U postgres -d smc_sellerservice << EOF
TRUNCATE companies CASCADE;
EOF

# Применить фикстуры заново
make fixtures-load
```

**Примечание:** Фикстуры используют `ON CONFLICT DO UPDATE`, поэтому можно просто запустить `make fixtures-load` повторно без предварительного удаления данных.
