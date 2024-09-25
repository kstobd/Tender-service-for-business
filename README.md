# Проект по Управлению Тендером

Этот проект представляет собой API для управления тендерами. Он включает в себя создание и редактирование тендеров, а также работу с организациями и ответственными за них сотрудниками.

## Установка под macOS

### Предварительные требования

- У вас должны быть установлены переменные окружения `SERVER_ADDRESS` и `POSTGRES_CONN` 
- PostgreSQL база данных должна быть запущена по указанному адресу.
- Гарантирована работа с Docker версии 20.10.12

### Переменные окружения

- `SERVER_ADDRESS` — Адрес и порт, на котором будет работать HTTP сервер. Пример: `0.0.0.0:8080`.
- `POSTGRES_CONN` — URL-строка для подключения к PostgreSQL в формате `postgres://{username}:{password}@{host}:{5432}/{dbname}`.

## Сборка и запуск проекта

### Сборка проекта из Dockerfile

```bash
sudo docker build -t avito_project .
```

### Запуск контейнера

```bash
docker run -p 8080:8080 -e SERVER_ADDRESS -e POSTGRES_CONN avito_project
```

## Использование API

Все эндпоинты начинаются с префикса `/api`.

### Пример запроса

#### Пинг эндпоинт

**Запрос:**

```
GET /api/ping
```

**Ответ:**

```
200 OK
Body: "ok"
```

## Примеры данных

### Таблица `tender`

```plaintext
id                                   | name    | description        | service_type | organization_id                         | creator_id                            | status  | created_at                   | updated_at
-------------------------------------|---------|--------------------|--------------|----------------------------------------|--------------------------------------|---------|------------------------------|------------------------------
94595083-d71a-442c-b112-a1407bdc5560 | Тендер 1 | Описание тендера   | Construction | 4c0e4b19-4206-42ea-a4d2-e4a07af0cbed | 0eeec920-40e3-4889-8913-f7802f5210e9 | CREATED | 2024-09-14 05:20:00.423492 | 2024-09-14 05:20:00.423492
```

### Таблица `organization_responsible`

```plaintext
id                                   | name    | description        | type | created_at                   | updated_at
-------------------------------------|---------|--------------------|------|------------------------------|------------------------------
4c0e4b19-4206-42ea-a4d2-e4a07af0cbed | TechCorp | IT Solutions Provider | IE   | 2024-09-14 04:54:56.583759 | 2024-09-14 04:54:56.583759
```

### Таблица `organization`

```plaintext
id                                   | name    | description        | type | created_at                   | updated_at
-------------------------------------|---------|--------------------|------|------------------------------|------------------------------
4c0e4b19-4206-42ea-a4d2-e4a07af0cbed | TechCorp | IT Solutions Provider | IE   | 2024-09-14 04:54:56.583759 | 2024-09-14 04:54:56.583759
```

## Тестирование

Для тестирования API вы можете использовать [Postman](https://www.postman.com).

---
