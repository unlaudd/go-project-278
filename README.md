Проект: Сокращатель ссылок

### Hexlet tests and linter status:
[![Actions Status](https://github.com/unlaudd/go-project-278/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/unlaudd/go-project-278/actions)
[![CI](https://github.com/unlaudd/go-project-278/actions/workflows/ci.yml/badge.svg)](https://github.com/unlaudd/go-project-278/actions)

### Деплой:
[https://url-shortener-452x.onrender.com](https://url-shortener-452x.onrender.com)

### Мониторинг ошибок:
[Sentry](https://sentry.io)

## API

**Базовый адрес:** `https://url-shortener-452x.onrender.com`

| Метод | Путь | Описание | Успешный ответ |
|-------|------|----------|---------------|
| `POST` | `/api/links` | Создать ссылку | `201 Created` + JSON |
| `GET`  | `/api/links` | Список ссылок | `200 OK` + [JSON] |
| `GET`  | `/api/links/:id` | Получить по ID | `200 OK` + JSON |
| `PUT`  | `/api/links/:id` | Обновить ссылку | `200 OK` + JSON |
| `DELETE` | `/api/links/:id` | Удалить ссылку | `204 No Content` |
| `GET` | `/r/:shortName` | Редирект на оригинал | `301 Moved` → Location |


**Пример создания ссылки:**
```bash
curl -X POST https://url-shortener-452x.onrender.com/api/links \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://example.com"}'
  ```

### Пагинация

Эндпоинт `GET /api/links` поддерживает пагинацию через параметр `range`:
`GET /api/links?range=[start,end]`


- `start`, `end` — инклюзивные границы (оба включаются)
- Если параметр не указан — используется дефолт `[0,9]` (первые 10 записей)

**Заголовок ответа:**

Content-Range: links start-end/total

**Примеры:**
```bash
# Первые 10 записей
curl -g "https://.../api/links?range=[0,9]"

# Записи 10-19
curl -g "https://.../api/links?range=[10,19]"

# Без пагинации (дефолт [0,9])
curl "https://.../api/links"
```

## Локальная разработка

### Требования
- Go ≥ 1.24
- Node.js ≥ 20 (LTS)
- PostgreSQL (локально или через Docker)

### Запуск
```bash
# Установите зависимости
go mod download
npm install
```

# Запустите фронтенд и бэкенд одновременно
```bash
npm run dev
```

Фронтенд: http://localhost:5173

API: http://localhost:8080
