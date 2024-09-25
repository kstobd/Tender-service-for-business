# Проект по Управлению Тендером

Этот проект представляет собой API для управления тендерами. Он включает в себя создание и редактирование тендеров, а также работу с организациями и ответственными за них сотрудниками.

## Установка под macOS

### Предварительные требования

- У вас должны быть установлены переменные окружения `SERVER_ADDRESS` и `POSTGRES_CONN` 
- PostgreSQL база данных должна быть запущена по указанному адресу.
- Docker: Рекомендуемая версия 20.10.12 или выше.

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

Это запустит ваш сервер на порту по адресу `SERVER_ADDRESS` с портом 8080.

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

### Таблица `employee`

```plaintext
                  id                  |  username  | first_name | last_name |         created_at         |         updated_at         
--------------------------------------+------------+------------+-----------+----------------------------+----------------------------
 0eeec920-40e3-4889-8913-f7802f5210e9 | test_user  | test       | user      | 2024-09-14 04:28:51.86996  | 2024-09-14 04:28:51.86996
 17f7cc8c-4b2b-4713-b9f4-3f5179548f18 | test_user2 | test2      | user2     | 2024-09-19 12:21:22.472431 | 2024-09-19 12:21:22.472431
(2 rows)
```


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

### 1. Создание нового тендера
Эндпоинт: `POST /api/tenders/new`

URL для тестирования: `http://localhost:8080/api/tenders/new`

Тело запроса (JSON):
```
{
  "name": "Новый тендер",
  "description": "Описание нового тендера",
  "serviceType": "Construction",
  "organizationId": "4c0e4b19-4206-42ea-a4d2-e4a07af0cbed",
  "creatorUsername": "test_user"
}
```
Ответ (успех):
```
{
    "createdAt": "2024-09-25T05:07:36Z",
    "description": "Описание нового тендера",
    "id": "3021a4d9-4dd3-429c-9c1b-f75f49a71883",
    "name": "Новый тендер",
    "serviceType": "Construction",
    "status": "CREATED",
    "version": 1
}
```

<img width="1019" alt="image" src="https://github.com/user-attachments/assets/1401f3c2-e258-434d-af73-a02210d6b525">

Если пользователь не найден
`401 Unauthorized` : User not found

<img width="1008" alt="image" src="https://github.com/user-attachments/assets/4ba51f22-1d76-4546-9bd5-0fe11aeaa7c8">

Пользователь не имеет доступ к организации
`403 Forbidden` : Недостаточно прав для выполнения действия.

<img width="1009" alt="image" src="https://github.com/user-attachments/assets/0fe958fb-6a79-49f8-9827-f8bd935602bf">

---

### 2. Список тендеров

**URL:** `http://localhost:8080/api/tenders`

Список тендеров с возможностью фильтрации по типу услуг.

- Если фильтры не заданы, возвращаются все тендеры.
  
  **Успешный вывод всех тендеров:**
  ![image](https://github.com/user-attachments/assets/764376ab-44f4-440c-a280-a183d7b643d6)

  **Применение фильтров:**
  `http://localhost:8080/api/tenders?limit=5&offset=0&service_type=Construction`

  ![image](https://github.com/user-attachments/assets/faa85178-fa18-44d0-9921-1088b9914bd3)

---

### 3. Тендеры пользователя

**URL:** `http://localhost:8080/api/tenders/my?username=test_user`

Получение списка тендеров текущего пользователя.

- Пример успешного ответа:
  ![image](https://github.com/user-attachments/assets/8b7fd362-d90d-41ab-b75c-4672224cb690)

- Если пользователь не найден:
  ![image](https://github.com/user-attachments/assets/9cca7d1d-004f-4fbd-bc6e-cfe9c37028c6)



### 4. Получение статуса тендера

**URL:** `http://localhost:8080/api/tenders/{tenderId}/status?username=test_user`

- Успешное получение статуса тендера:
  ![image](https://github.com/user-attachments/assets/bbd17732-69da-4216-8b84-6cf2b4079fcb)

- Ошибка при несуществующем пользователе:
  ![image](https://github.com/user-attachments/assets/425539bc-35f1-407a-8e38-1a913f6f7c6c)

- Ошибка при несуществующем тендере:
  ![image](https://github.com/user-attachments/assets/66805b05-ac44-404f-a3fd-3be5f084770e)

---

### 5. Изменение статуса тендера

**Эндпоинт:** `PUT /tenders/{tenderId}/status`

**URL:** `http://localhost:8080/api/tenders/3021a4d9-4dd3-429c-9c1b-f75f49a71883/status?status=Published&username=test_user`

- Пример успешного ответа:
  ![image](https://github.com/user-attachments/assets/a930de05-aab6-4515-b1ab-6bad89b1a465)

- Проверка всех тендеров после изменения статуса:
  ![image](https://github.com/user-attachments/assets/701e8844-f31e-42d3-af1a-ad5f69153baa)

---

### 6. Изменение тендера

**Эндпоинт:** `PATCH /tenders/{tenderId}/edit`

**URL:** `http://localhost:8080/api/tenders/3021a4d9-4dd3-429c-9c1b-f75f49a71883/edit?username=test_user`

**Тело запроса:**

```json
{
  "name": "Новое имя для тендера",
  "description": "Новое описание для тендера",
  "serviceType": "Construction"
}
```


<img width="997" alt="image" src="https://github.com/user-attachments/assets/c6f6f62d-59e6-41f1-967f-29492362253d">

Так же присутствуют все стандартные проверки

---

### Пример логов работы API
Хочу заметить как все действия замечательно логгируются

```plaintext
2024/09/25 02:33:33 GetTenderStatusHandler: Getting status for tender 3021a4d9-4dd3-429c-9c1b-f75f49a71883
2024/09/25 02:34:26 GetTenderStatusHandler: Successfully retrieved status in 1.845708ms
2024/09/25 02:37:21 GetTendersHandler: Retrieving list of tenders
2024/09/25 02:37:21 GetTendersHandler: Successfully retrieved tenders in 1.397042ms
```

Так же реализованы 
```
/api/tenders/{tenderId}/rollback/{version} Methods("PUT")

/api/bids/new Methods("POST")
/api/bids/my Methods("GET")
/api/bids/{tenderId}/list Methods("GET")
/api/bids/{bidId}/submit_decision Methods("PUT")
```
