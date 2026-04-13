# Проект: Сокращатель ссылок

[![Actions Status](https://github.com/unlaudd/go-project-278/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/unlaudd/go-project-278/actions)
[![CI](https://github.com/unlaudd/go-project-278/actions/workflows/ci.yml/badge.svg)](https://github.com/unlaudd/go-project-278/actions)

Сервис для сокращения длинных ссылок с веб-интерфейсом, аналитикой посещений и удобным API.

---

## Деплой

- **Приложение**: [https://url-shortener-452x.onrender.com](https://url-shortener-452x.onrender.com)
- **Мониторинг ошибок**: [Sentry](https://sentry.io)

---

## Стек технологий

| Компонент | Технология |
|-----------|-----------|
| Бэкенд | Go 1.24, Gin, sqlc, lib/pq |
| Фронтенд | React, Vite (пакет `@hexlet/project-url-shortener-frontend`) |
| База данных | PostgreSQL, миграции через goose |
| Деплой | Docker (multi-stage), Render, Caddy |
| Тестирование | testify, golangci-lint |
| Мониторинг | Sentry |

---

## Быстрый старт (локально)

### Требования

- Go ≥ 1.24
- Node.js ≥ 20 (LTS)
- PostgreSQL (локально или в Docker)

### Установка и запуск

```bash
# 1. Клонируйте репозиторий
git clone <repository-url>
cd go-project-278
```

# 2. Установите зависимости бэкенда и фронтенда
```bash
go mod download
npm install
```

# 3. Запустите приложение в режиме разработки
```bash
npm run dev
```

После запуска:

    Фронтенд: http://localhost:5173
    API: http://localhost:8080

    CORS настроен для http://localhost:5173 — запросы с фронтенда на бэкенд проходят без дополнительных настроек.

# 4. Переменные окружения

| Переменная   |          Описание                                      |                      Пример                            |
|--------------|--------------------------------------------------------|--------------------------------------------------------|	
| DATABASE_URL | Строка подключения к PostgreSQL                        | postgres://user:pass@localhost:5432/db?sslmode=disable |
| BASE_URL     | Базовый адрес приложения для генерации коротких ссылок | http://localhost:8080                                  |
| BACKEND_PORT | Порт, на котором слушает бэкенд                        | 8080                                                   |
| ENVIRONMENT  | Окружение для Sentry (development/production)          | development                                            |

# 5. API
Базовый адрес: http://localhost:8080 (локально) или https://url-shortener-452x.onrender.com (продакшен)

## Управление ссылками

| Метод        |    Путь          |       Описание        |   Успешный ответ   |
|--------------|------------------|-----------------------|--------------------|
| POST         |  /api/links      | Создать ссылку        | 201 Created + JSON |
| GET          |  /api/links      | Список ссылок         | 200 OK + [JSON]    |
| GET          |  /api/links/:id  | Получить ссылку по ID | 200 OK + JSON      |
| PUT          |  /api/links/:id  | Обновить ссылку       | 200 OK + JSON      |
| DELETE       |  /api/links/:id  | Удалить ссылку        | 204 No Content     |

Пример создания ссылки:
```bash
curl -X POST http://localhost:8080/api/links \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://example.com"}'
```
Ответ:
```json
{
  "id": 1,
  "original_url": "https://example.com",
  "short_name": "abc123",
  "short_url": "http://localhost:8080/r/abc123"
}
```

## Пагинация
Эндпоинт GET /api/links поддерживает пагинацию через параметр range:
```bash
GET /api/links?range=[start,end]
```

  - start, end — инклюзивные границы (обе включаются в результат)
  - Если параметр не указан — используется дефолт [0,9] (первые 10 записей)

Заголовок ответа:

```bash
Content-Range: links start-end/total
```

Примеры:
```bash
# Первые 10 записей
curl -g "http://localhost:8080/api/links?range=[0,9]"

# Записи 10-19
curl -g "http://localhost:8080/api/links?range=[10,19]"

# Без параметра (дефолт [0,9])
curl "http://localhost:8080/api/links"
```

# 6. Аналитика посещений

| Метод |      Путь        |          Описание                    |
|-------|------------------|--------------------------------------|
| GET   | /r/:code         | Редирект на оригинал + запись визита |
| GET   | /api/link_visits | Список посещений с пагинацией        |

## Формат записи посещения

```json
{
  "id": 5,
  "link_id": 1,
  "created_at": "2025-10-31T13:01:43Z",
  "ip": "172.18.0.1",
  "user_agent": "curl/8.5.0",
  "referer": "https://example.com",
  "status": 301
}
```
Пример запроса:

```bash
curl -g "http://localhost:8080/api/link_visits?range=[0,9]"
```

Ответ:

```bash
Content-Range: link_visits 0-9/42
[
  {"id":1, "link_id":1, "ip":"...", "status":301, ...}
]
```

# 7. Тестирование

```bash
# Запустить все тесты
make test

# Запустить тесты с детектором гонок
make test-race

# Запустить линтер
make lint

# Автоматически исправить ошибки линтера (где возможно)
make lint-fix
```

# 8. Полезные команды Makefile

```bash
make help     # Показать список доступных команд
make run      # Запустить приложение через go run
make build    # Скомпилировать бинарник
make clean    # Удалить скомпилированный бинарник
make fmt      # Отформатировать код через goimports
make deps     # Обновить зависимости
make dev      # Запустить с авто-перезагрузкой (air)
```
