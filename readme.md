

### Тестирование API

#### 1. Регистрация пользователя

- Метод: POST
- URL: `{{base_url}}/api/auth/register`
- Headers:
  - Content-Type: application/json
- Body (raw, JSON):
```json
{
    "login": "test_user",
    "email": "user@example.com",
    "password": "test123"
}
```

#### 2. Авторизация пользователя

- Метод: POST
- URL: `{{base_url}}/api/auth/login`
- Headers:
  - Content-Type: application/json
- Body (raw, JSON):
```json
{
    "login": "test_user",
    "email": "user@example.com",
    "password": "test123"
}
```

#### 3. Получение информации о текущем пользователе

- Метод: GET
- URL: `{{base_url}}/api/auth/me`
- Authorization:
  - Type: Bearer Token
  - Token: `{{token}}`

#### 4. Обновление описания профиля

- Метод: PUT
- URL: `{{base_url}}/api/profile`
- Authorization:
  - Type: Bearer Token
  - Token: `{{token}}`
- Headers:
  - Content-Type: application/json
- Body (raw, JSON):
```json
{
    "description": "Это мой тестовый профиль"
}
```

#### 5. Загрузка изображения профиля

- Метод: POST
- URL: `{{base_url}}/api/profile/image`
- Authorization:
  - Type: Bearer Token
  - Token: `{{token}}`
- Headers:
  - Content-Type: application/json
- Body (raw, JSON):
```json
{
    "image": "data:image/jpeg;base64,/9j/4AAQSkZJRgAB..."
}
```

Для добавления изображения в base64:
1. В Postman нажмите в поле ввода JSON правой кнопкой и выберите "Insert file as base64"
2. Выберите файл изображения
3. Отредактируйте вставленный текст, добавив перед ним `data:image/jpeg;base64,`

#### 5. Загрузка изображения профиля

- Метод: POST
- URL: `{{base_url}}/api/profile/banner`
- Authorization:
  - Type: Bearer Token
  - Token: `{{token}}`
- Headers и Body аналогично загрузке изображения профиля

#### 7. Создание новой карточки

- Метод: POST
- URL: `{{base_url}}/api/cards`
- Authorization:
  - Type: Bearer Token
  - Token: `{{token}}`
- Headers:
  - Content-Type: application/json
- Body (raw, JSON):
```json
{
    "title": "Моя первая карточка",
    "description": "Это тестовое описание карточки",
    "text": "Здесь может быть подробный текст карточки.",
    "image": "data:image/jpeg;base64,/9j/4AAQSkZJRgAB..."
}
```

После успешного создания карточки:
1. Во вкладке "Tests" добавьте:
```javascript
var jsonData = pm.response.json();
pm.environment.set("cardId", jsonData.id);
```

#### 8. Получение списка всех карточек

- Метод: GET
- URL: `{{base_url}}/api/cards?page=1&limit=4`

#### 9. Получение карточки по ID

- Метод: GET
- URL: `{{base_url}}/api/cards/{{cardId}}`

#### 10. Получение карточек пользователя

- Метод: GET
- URL: `{{base_url}}/api/cards/user/{{userId}}?page=1&limit=4`

#### 11. Обновление карточки

- Метод: PUT
- URL: `{{base_url}}/api/cards/{{cardId}}`
- Authorization:
  - Type: Bearer Token
  - Token: `{{token}}`
- Headers:
  - Content-Type: application/json
- Body (raw, JSON):
```json
{
    "title": "Обновленная карточка",
    "description": "Обновленное описание карточки",
    "text": "Здесь обновленный текст карточки."
}
```

#### 12. Лайк карточки

- Метод: POST
- URL: `{{base_url}}/api/cards/{{cardId}}/like`
- Authorization:
  - Type: Bearer Token
  - Token: `{{token}}`

#### 13. Удаление лайка с карточки

- Метод: DELETE
- URL: `{{base_url}}/api/cards/{{cardId}}/like`
- Authorization:
  - Type: Bearer Token
  - Token: `{{token}}`

#### 14. Удаление карточки

- Метод: DELETE
- URL: `{{base_url}}/api/cards/{{cardId}}`
- Authorization:
  - Type: Bearer Token
  - Token: `{{token}}`

### Пошаговый план тестирования всего API

Для полного тестирования API выполните следующие шаги в указанном порядке:

1. **Подготовка**:
   - Запустите сервер: `go run cmd/main.go`
   - Настройте переменные окружения в Postman
   - Создайте коллекцию и папки для запросов

2. **Аутентификация**:
   - Выполните запрос регистрации пользователя
   - Проверьте ответ: должен вернуться токен и данные пользователя
   - Выполните запрос авторизации (логина)
   - Проверьте, что токен сохранен в переменной окружения
   - Выполните запрос для получения информации о текущем пользователе

3. **Работа с профилем**:
   - Обновите описание профиля
   - Загрузите изображение профиля (подготовьте файл заранее)
   - Загрузите баннер профиля (подготовьте файл заранее)
   - Проверьте обновленный профиль через запрос `/api/auth/me`

4. **Работа с карточками**:
   - Создайте новую карточку (с изображением или без)
   - Убедитесь, что ID карточки сохранен в переменной `cardId`
   - Получите список всех карточек и проверьте наличие созданной
   - Получите карточку по ID
   - Получите карточки текущего пользователя
   - Обновите карточку
   - Поставьте лайк карточке
   - Проверьте, что количество лайков увеличилось
   - Удалите лайк с карточки
   - Проверьте, что количество лайков уменьшилось

5. **Проверка взаимодействия между пользователями** (опционально):
   - Создайте второго пользователя (в новом окружении)
   - Авторизуйтесь под вторым пользователем
   - Просмотрите карточки первого пользователя
   - Поставьте лайк карточке первого пользователя
   - Вернитесь к первому пользователю и проверьте лайки

6. **Удаление данных** (по желанию):
   - Удалите созданную карточку
   - Проверьте, что карточка удалена, запросив список всех карточек

### Советы по тестированию в Postman

1. **Сохранение запросов:** После создания запроса нажмите "Save" для добавления в коллекцию.

2. **Организация запросов:**
   - Правой кнопкой на коллекции выберите "Add Folder"
   - Создайте папки: Auth, Profile, Cards
   - Распределите запросы по соответствующим папкам

3. **Экспорт коллекции:**
   - Для совместного использования нажмите "..." рядом с названием коллекции
   - Выберите "Export" и сохраните JSON-файл
   - Другие пользователи могут импортировать эту коллекцию

4. **Запуск коллекции:**
   - Нажмите "..." рядом с названием коллекции и выберите "Run collection"
   - Выберите порядок запросов и настройки 
   - Запустите всю последовательность запросов одним кликом

5. **Проверка результатов:**
   - Используйте Tests для автоматической проверки результатов
   - Пример скрипта для проверки успешного ответа:
   ```javascript
   pm.test("Ответ содержит код 200", function () {
       pm.response.to.have.status(200);
   });
   ```

6. **Отладка проблем:**
   - Включите подробный вывод в консоль Postman
   - Проверяйте ответы сервера на каждый запрос
   - Для ошибок с изображениями проверьте формат base64 и добавьте правильный префикс

### Создание карточки

```
POST http://localhost:4000/api/cards
```

#### Параметры

Отправляется как `multipart/form-data`:

| Поле | Тип | Описание |
|------|-----|----------|
| title | string | Заголовок карточки (обязательно) |
| description | string | Краткое описание |
| text | string | Текст карточки |
| image | file | Изображение карточки (опционально) |

#### Пример запроса

```
curl -X POST http://localhost:4000/api/cards \
  -H "Authorization: Bearer TOKEN" \
  -F "title=Моя новая карточка" \
  -F "description=Описание карточки" \
  -F "text=Полный текст карточки" \
  -F "image=@/путь/к/изображению.jpg"
```

### Обновление профиля

```
PUT http://localhost:4000/api/profile
```

#### Параметры

Отправляется как JSON:

| Поле | Тип | Описание |
|------|-----|----------|
| description | string | Описание профиля |

#### Пример запроса

```
curl -X PUT http://localhost:4000/api/profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TOKEN" \
  -d '{
    "description": "Обновленное описание профиля"
  }'
```

### Загрузка изображения профиля

```
POST http://localhost:4000/api/profile/image
```

#### Параметры

Отправляется как `multipart/form-data`:

| Поле | Тип | Описание |
|------|-----|----------|
| image | file | Файл изображения |

#### Пример запроса

```
curl -X POST http://localhost:4000/api/profile/image \
  -H "Authorization: Bearer TOKEN" \
  -F "image=@/путь/к/изображению.jpg"
```

### Загрузка баннера профиля

```
POST http://localhost:4000/api/profile/banner
```

#### Параметры

Отправляется как `multipart/form-data`:

| Поле | Тип | Описание |
|------|-----|----------|
| image | file | Файл изображения (макс. 1200x400) |

#### Пример запроса

```
curl -X POST http://localhost:4000/api/profile/banner \
  -H "Authorization: Bearer TOKEN" \
  -F "image=@/путь/к/баннеру.jpg"
```

### Обновление карточки

```
PUT http://localhost:4000/api/cards/:cardId
```

#### Параметры

Отправляется как `multipart/form-data`:

| Поле | Тип | Описание |
|------|-----|----------|
| title | string | Заголовок карточки (обязательно) |
| description | string | Краткое описание |
| text | string | Текст карточки |
| image | file | Новое изображение карточки (опционально) |
| remove_image | string | Если установлено в "true", удаляет текущее изображение |

#### Пример запроса

```
curl -X PUT http://localhost:4000/api/cards/123 \
  -H "Authorization: Bearer TOKEN" \
  -F "title=Обновленный заголовок" \
  -F "description=Новое описание" \
  -F "text=Новый текст карточки" \
  -F "image=@/путь/к/изображению.jpg"
```

Для удаления изображения:

```
curl -X PUT http://localhost:4000/api/cards/123 \
  -H "Authorization: Bearer TOKEN" \
  -F "title=Заголовок" \
  -F "remove_image=true"
```

## Тестирование через Postman

### Подготовка

1. Запустите сервер
2. Создайте новую коллекцию в Postman с названием "Социальная сеть с карточками"
3. Создайте переменные окружения:
   - `base_url` = `http://localhost:4000`
   - `token` - будет заполнена автоматически после авторизации
   - `userId` - будет заполнена автоматически после авторизации

### Настройка авторизации

После успешного логина добавьте в раздел Tests следующий скрипт для автоматического сохранения токена:

```javascript
if (pm.response.code === 200) {
    var jsonData = pm.response.json();
    pm.environment.set("token", jsonData.token);
    pm.environment.set("userId", jsonData.user.id);
}
```

### Тестирование API

#### Регистрация пользователя

- Method: POST
- URL: {{base_url}}/api/auth/register
- Body (JSON):
  ```json
  {
    "login": "testuser",
    "password": "password123"
  }
  ```

#### Вход в систему

- Method: POST
- URL: {{base_url}}/api/auth/login
- Body (JSON):
  ```json
  {
    "login": "testuser",
    "password": "password123"
  }
  ```

#### Обновление профиля

- Method: PUT
- URL: {{base_url}}/api/profile
- Auth: Bearer Token: {{token}}
- Body (JSON):
  ```json
  {
    "description": "Мой тестовый профиль"
  }
  ```

#### Загрузка изображения профиля

- Method: POST
- URL: {{base_url}}/api/profile/image
- Auth: Bearer Token: {{token}}
- Body (form-data):
  - Key: image
  - Type: File
  - Value: выберите файл с изображением

#### Загрузка баннера профиля

- Method: POST
- URL: {{base_url}}/api/profile/banner
- Auth: Bearer Token: {{token}}
- Body (form-data):
  - Key: image
  - Type: File
  - Value: выберите файл с изображением для баннера

#### Создание карточки

- Method: POST
- URL: {{base_url}}/api/cards
- Auth: Bearer Token: {{token}}
- Body (form-data):
  - Key: title, Type: Text, Value: Моя карточка
  - Key: description, Type: Text, Value: Описание
  - Key: text, Type: Text, Value: Полный текст
  - Key: image, Type: File, Value: выберите файл с изображением