Проект: Сокращатель ссылок

### Hexlet tests and linter status:
[![Actions Status](https://github.com/unlaudd/go-project-278/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/unlaudd/go-project-278/actions)
[![CI](https://github.com/unlaudd/go-project-278/actions/workflows/ci.yml/badge.svg)](https://github.com/unlaudd/go-project-278/actions)

### Деплой:
[https://url-shortener-452x.onrender.com](https://url-shortener-452x.onrender.com)

### Мониторинг ошибок:
[Sentry](https://sentry.io)

## 📡 API

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