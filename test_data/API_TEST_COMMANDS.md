# API Test Commands

Набор curl команд для тестирования SMC-SellerService API согласно OpenAPI спецификации.

---

## Companies API

### 1. Создание компании (POST /api/v1/companies)
**Описание:** Создает новую компанию. Доступно только superuser.
**Ожидаемый результат:** 201 Created, возвращается созданная компания с ID
**Данные:** Полная информация о компании (название, адреса, рабочие часы, менеджеры)

```bash
curl -X POST http://localhost:8081/api/v1/companies \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 123456789" \
  -H "X-User-Role: superuser" \
  -d @test_data/create_company.json | jq .
```

---

### 2. Создание компании от имени обычного user (должен вернуть 403)
**Описание:** Попытка создать компанию от имени обычного пользователя
**Ожидаемый результат:** 403 Forbidden, `{"code":403,"message":"access denied"}`
**Данные:** Те же данные, но с ролью "user"

```bash
curl -X POST http://localhost:8081/api/v1/companies \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 123" \
  -H "X-User-Role: user" \
  -d @test_data/create_company.json | jq .
```

---

### 3. Получение списка компаний (GET /api/v1/companies)
**Описание:** Возвращает список всех компаний. Публичный endpoint, авторизация не требуется
**Ожидаемый результат:** 200 OK, массив компаний с полной информацией
**Данные:** Нет (опционально query параметры для фильтрации)

```bash
curl http://localhost:8081/api/v1/companies | jq .
```

---

### 4. Получение компании по ID (GET /api/v1/companies/{id})
**Описание:** Возвращает детальную информацию о компании. Публичный endpoint
**Ожидаемый результат:** 200 OK, объект компании
**Данные:** Нет

```bash
curl http://localhost:8081/api/v1/companies/1 | jq .
```

---

### 5. Получение несуществующей компании (должен вернуть 404)
**Описание:** Запрос компании с несуществующим ID
**Ожидаемый результат:** 404 Not Found, `{"code":404,"message":"company not found"}`
**Данные:** Нет

```bash
curl http://localhost:8081/api/v1/companies/999 | jq .
```

---

### 6. Обновление компании от имени superuser (PUT /api/v1/companies/{id})
**Описание:** Обновляет данные компании. Доступно superuser или менеджеру компании
**Ожидаемый результат:** 200 OK, обновленная компания
**Данные:** Частичное обновление (только измененные поля)

```bash
curl -X PUT http://localhost:8081/api/v1/companies/1 \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 123456789" \
  -H "X-User-Role: superuser" \
  -d @test_data/update_company.json | jq .
```

---

### 7. Обновление компании от имени не-менеджера (должен вернуть 403)
**Описание:** Попытка обновить компанию пользователем, который не является менеджером
**Ожидаемый результат:** 403 Forbidden, `{"code":403,"message":"access denied"}`
**Данные:** Те же данные, но с user_id, не являющимся менеджером

```bash
curl -X PUT http://localhost:8081/api/v1/companies/1 \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 999999" \
  -H "X-User-Role: user" \
  -d @test_data/update_company.json
```

---

### 8. Удаление компании от имени superuser (DELETE /api/v1/companies/{id})
**Описание:** Удаляет компанию. Доступно только superuser
**Ожидаемый результат:** 204 No Content (пустой ответ)
**Данные:** Нет

```bash
curl -X DELETE http://localhost:8081/api/v1/companies/2 \
  -H "X-User-ID: 123456789" \
  -H "X-User-Role: superuser" \
  -w "\nHTTP Status: %{http_code}\n"
```

---

### 9. Удаление компании от имени обычного user (должен вернуть 403)
**Описание:** Попытка удалить компанию от имени обычного пользователя
**Ожидаемый результат:** 403 Forbidden
**Данные:** Нет

```bash
curl -X DELETE http://localhost:8081/api/v1/companies/2 \
  -H "X-User-ID: 999999" \
  -H "X-User-Role: user"
```

---

## Фильтрация

### 10. Фильтрация по тегам (GET /api/v1/companies?tags=...)
**Описание:** Возвращает компании, у которых есть указанные теги
**Ожидаемый результат:** 200 OK, отфильтрованный список компаний
**Данные:** Query параметр `tags` (можно несколько через запятую)

```bash
curl "http://localhost:8081/api/v1/companies?tags=%23мойка" | jq '.companies[] | {id, name, tags}'
```

---

### 11. Фильтрация по городу (GET /api/v1/companies?city=...)
**Описание:** Возвращает компании в указанном городе
**Ожидаемый результат:** 200 OK, отфильтрованный список компаний
**Данные:** Query параметр `city`

```bash
curl "http://localhost:8081/api/v1/companies?city=Москва" | jq '.companies[] | {id, name}'
```

---

## Services API

### 12. Создание услуги (POST /api/v1/companies/{company_id}/services)
**Описание:** Создает новую услугу для компании. Доступно superuser или менеджеру компании
**Ожидаемый результат:** 201 Created, созданная услуга с ID
**Данные:** Название, описание, длительность, адреса где оказывается услуга

```bash
curl -X POST http://localhost:8081/api/v1/companies/1/services \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 123456789" \
  -H "X-User-Role: superuser" \
  -d @test_data/create_service.json | jq .
```

---

### 13. Получение списка услуг компании (GET /api/v1/companies/{company_id}/services)
**Описание:** Возвращает все услуги указанной компании. Публичный endpoint
**Ожидаемый результат:** 200 OK, массив услуг
**Данные:** Нет

```bash
curl http://localhost:8081/api/v1/companies/1/services | jq .
```

---

### 14. Получение услуги по ID (GET /api/v1/companies/{company_id}/services/{service_id})
**Описание:** Возвращает детальную информацию об услуге. Публичный endpoint
**Ожидаемый результат:** 200 OK, объект услуги
**Данные:** Нет

```bash
curl http://localhost:8081/api/v1/companies/1/services/1 | jq .
```

---

### 15. Обновление услуги (PUT /api/v1/companies/{company_id}/services/{service_id})
**Описание:** Обновляет данные услуги. Доступно superuser или менеджеру компании
**Ожидаемый результат:** 200 OK, обновленная услуга
**Данные:** Частичное обновление (только измененные поля)

```bash
curl -X PUT http://localhost:8081/api/v1/companies/1/services/1 \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 123456789" \
  -H "X-User-Role: superuser" \
  -d @test_data/update_service.json | jq .
```

---

### 16. Удаление услуги (DELETE /api/v1/companies/{company_id}/services/{service_id})
**Описание:** Удаляет услугу. Доступно superuser или менеджеру компании
**Ожидаемый результат:** 204 No Content
**Данные:** Нет

```bash
curl -X DELETE http://localhost:8081/api/v1/companies/1/services/1 \
  -H "X-User-ID: 123456789" \
  -H "X-User-Role: superuser" \
  -w "\nHTTP Status: %{http_code}\n"
```

---

## Тестовые данные

### test_data/create_company.json
```json
{
  "name": "Автомойка Премиум",
  "logo": "https://storage.example.com/logos/company-123.png",
  "description": "Профессиональная автомойка и детейлинг в центре Москвы",
  "tags": ["#мойка", "#детейлинг", "#москва", "#премиум"],
  "addresses": [{
    "city": "Москва",
    "street": "Тверская улица",
    "building": "10к1",
    "coordinates": {"latitude": 55.755826, "longitude": 37.617299}
  }],
  "working_hours": {
    "monday": {"isOpen": true, "openTime": "09:00", "closeTime": "21:00"},
    "tuesday": {"isOpen": true, "openTime": "09:00", "closeTime": "21:00"},
    "wednesday": {"isOpen": true, "openTime": "09:00", "closeTime": "21:00"},
    "thursday": {"isOpen": true, "openTime": "09:00", "closeTime": "21:00"},
    "friday": {"isOpen": true, "openTime": "09:00", "closeTime": "21:00"},
    "saturday": {"isOpen": true, "openTime": "10:00", "closeTime": "20:00"},
    "sunday": {"isOpen": false}
  },
  "manager_ids": [123456789]
}
```

### test_data/update_company.json
```json
{
  "name": "Автомойка Премиум Плюс",
  "description": "Обновленное описание: Лучшая автомойка в Москве!"
}
```

### test_data/create_service.json
```json
{
  "name": "Комплексная мойка",
  "description": "Полная мойка кузова, дисков, ковриков и салона",
  "average_duration": 60,
  "address_ids": [1]
}
```

### test_data/update_service.json
```json
{
  "name": "Комплексная мойка Premium",
  "average_duration": 90
}
```

---

## Примечания

1. **Заголовки авторизации** (для protected endpoints):
   - `X-User-ID` - ID пользователя (int64)
   - `X-User-Role` - роль: `superuser` или `user`

2. **Роли:**
   - `superuser` - полный доступ ко всем операциям
   - `user` - может изменять только компании, где он указан в `manager_ids`

3. **Коды ответов:**
   - `200 OK` - успешное получение/обновление
   - `201 Created` - успешное создание
   - `204 No Content` - успешное удаление
   - `400 Bad Request` - невалидные данные
   - `401 Unauthorized` - отсутствуют заголовки авторизации
   - `403 Forbidden` - недостаточно прав
   - `404 Not Found` - ресурс не найден
   - `500 Internal Server Error` - внутренняя ошибка

4. **URL-кодирование:**
   - Символ `#` в query параметрах должен быть закодирован как `%23`
   - Пример: `?tags=%23мойка` вместо `?tags=#мойка`
