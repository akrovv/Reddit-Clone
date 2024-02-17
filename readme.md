# Клон реддита
![Build Status](https://github.com/akrovv/reddit-clone/actions/workflows/.github.yml/badge.svg)

## Клон реддита
```
Возможности:
  - Авторизация / Регистрация
  - Посмотреть посты с выбранной категорией
  - Опубликовать пост с сообщением / ссылкой на ресурс
  - Поставить лайк / дизлайк под постом
  - Удалить пост
  - Оставить комментарий под постом
  - Удалить комментарий
  - Зайти в профиль к пользователю и посмотреть все его опубликованные посты
```


## Инструкция по запуску
```
git clone https://github.com/akrovv/reddit-clone.git
docker-compose up
```

## Стек технологий
Backend: Golang, PostgreSQL, Mongo, Redis, Rest API, docker-compose. Frontend: JS.

## Особенности
- Различные БД: (Mongo для постов, PostgreSQL для пользователей, Redis для сессий)
- Casbin для ролей
- Mock сервисов для тестирования handler'ов
- CI-CD **(lint+test)**

## Makefile
- Make build - сборка проекта
- Make lint - linter 
- Make test-api - тестирование с покрытием
- Make test-permissions - тестирование ролей (работает только при запущенном приложении)

## Документация

Входные значения - **body**.  
Выходные значения - **что должен вернуть сервер**.

Структура поста:
* score - **рейтинг**
* views - **количество просмотров**
* type - **тип** (link - только ссылка на ресурс, text - текст к посту).
* title - **заголовок**
* author - **автор**. Содержит в себе имя пользователя (автора) и его id
* category - **категория**. Также доступны следующие категории: funny, videos, programming, news, fashion
* text - **текст** (если был выбран type=Text, в ином случае text="")
* votes - **голоса** всех, кто проголосовал за пост. Содержит в себе **id** пользователя и его **оценка (-1, 0, 1)**.
* comments - **комментарии**. Содержит в себе **дату публикации**, **автора**, **текст**, **id**
* created - **дата публикации**
* upvotePercentage - **рейтинг** в процентах
* id - **id**

Структура пользователя:
* username - **имя пользователя (логин)**
* password - **пароль**
* id - **id**

Структура токена:
* token - **JWT токен**

### Получить все посты GET /api/posts/

**Принимает: -**  
**Возвращает: массив в объектах в JSON post**

Пример возможного запроса:
```bash
curl -X 'GET' "http://localhost:8080/api/posts/" 
```

Пример успешного ответа, если есть посты:
```json
[
    {
        "score":1,
        "views":2,
        "type":"text",
        "title":"akroakro",
        "author":
        {
            "username":"sheldon cooper",
            "id":"18890d6a9ce0cdbb5a8ce0c1"
        },
        "category":"music",
        "text":"akroakro",
        "votes":
        [
            {
                "user":"18890d6a9ce0cdbb5a8ce0c1",
                "vote":1
            }
        ],
        "comments":[],
        "created":"2024-02-17T11:34:12.189Z",
        "upvotePercentage":100,
        "id":"a760045d-79ea-51db-b3a8-db5e785d0551"
    },
    ...
]
```

Если постов нет:
```json
[]
```

В случае возникновения ошибки - сервер выдает ошибку.
### Создать пост POST /api/posts
**Принимает: объект JSON post**  
**Возвращает: объект JSON post**  
**Требование: добавить пост может только авторизованный пользователь**

Пример успешного ответа:
```json
{
    "score": 1,
    "views": 0,
    "type": "text",
    "title": "hello world",
    "url": "",
    "author": {
        "username": "sheldon cooper",
        "id": "65d09ade1d06de00132f7eb1"
    },
    "category": "music",
    "text": "the big bang theory",
    "votes": [
        {
            "user": "65d09ade1d06de00132f7eb1",
            "vote": 1
        }
    ],
    "comments": [],
    "created": "2024-02-17T11:39:35.047Z",
    "upvotePercentage": 100,
    "id": "65d09af71d06de00132f7eb2"
}
```

В случае возникновения ошибки - сервер выдает ошибку.

### Получить посты по категории GET /api/posts/{CATEGORY_NAME}
**Принимает: -**  
**Возвращает: массив в объектах JSON post**

Пример возможного запроса:
```bash
curl -X 'GET' "http://localhost:8080/api/posts/music"
```

Пример успешного ответа:

```json
[
    {
        "score":1,
        "views":2,
        "type":"text",
        "title":"akroakro",
        "author":
        {
            "username":"sheldon cooper",
            "id":"18890d6a9ce0cdbb5a8ce0c1"
        },
        "category":"music",
        "text":"akroakro",
        "votes":
        [
            {
                "user":"18890d6a9ce0cdbb5a8ce0c1",
                "vote":1
            }
        ],
        "comments":[],
        "created":"2024-02-17T11:34:12.189Z",
        "upvotePercentage":100,
        "id":"a760045d-79ea-51db-b3a8-db5e785d0551"
    },
    ...
]
```

В случае возникновения ошибки - сервер выдает ошибку.
### Получить один пост GET /api/post/{POST_ID}

**Принимает: -**  
**Возвращает: объект JSON post**

Пример возможного запроса:
```bash
curl -X 'GET' "http://localhost:8080/api/post/id"
```

Пример успешного ответа:
```json
{
    "score": 1,
    "views": 5,
    "type": "text",
    "title": "Sheldon Cooper",
    "author": {
        "username": "sheldon cooper",
        "id": "65d09ade1d06de00132f7eb1"
    },
    "category": "music",
    "text": "the Big Bang theory",
    "votes": [
        {
            "user": "65d09ade1d06de00132f7eb1",
            "vote": 1
        }
    ],
    "comments": [
        {
            "created": "2024-02-17T11:51:04.898Z",
            "author": {
                "username": "akro",
                "id": "65d09ade1d06de00132f7eb1"
            },
            "body": "Good!",
            "id": "65d09da81d06de00132f7eb4"
        }
    ],
    "created": "2024-02-17T11:51:01.083Z",
    "upvotePercentage": 100,
    "id": "65d09da51d06de00132f7eb3"
}
```
В случае возникновения ошибки - сервер выдает ошибку.

### Удалить пост DELETE /api/post/{POST_ID}
**Принимает: -**  
**Возвращает: объект JSON post**  
**Требование: удалить пост может только авторизованный пользователь**

Пример возможного запроса:
```bash
curl -X 'DELETE' "http://localhost:8080/api/post/id"
```

Пример успешного ответа:
```json
{
    "message": "success"
}
```

### Получить посты пользователя GET /api/user/{USER_LOGIN}
**Принимает: -**  
**Возвращает: массив в объектах JSON post**

Пример возможного запроса:
```bash
curl -X 'GET' "http://localhost:8080/api/user/sheldon"
```

Пример успешного ответа:
```json
[
    {
        "score":1,
        "views":2,
        "type":"text",
        "title":"akroakro",
        "author":
        {
            "username":"sheldon cooper",
            "id":"18890d6a9ce0cdbb5a8ce0c1"
        },
        "category":"music",
        "text":"akroakro",
        "votes":
        [
            {
                "user":"18890d6a9ce0cdbb5a8ce0c1",
                "vote":1
            }
        ],
        "comments":[],
        "created":"2024-02-17T11:34:12.189Z",
        "upvotePercentage":100,
        "id":"a760045d-79ea-51db-b3a8-db5e785d0551"
    },
    ...
]
```

### Рейтинг поста GET /api/post/{POST_ID}/[upvote, downvote, unvote]
**Принимает: -**  
**Возвращает: объект JSON post**  
**Требование: проголосовать может только авторизованный пользователь**  
**Примечание: псле POST_ID/ может стоять только один из предложенных вариантов. Upvote - поднять рейтинг, downvote - опустить рейтинг, unvote - убрать голос**

Пример возможного запроса:
```bash
curl -X 'GET' "http://localhost:8080/api/post/id/upvote"
```

Пример успешного ответа:
```json
{
    "score": 1,
    "views": 5,
    "type": "text",
    "title": "Sheldon Cooper",
    "author": {
        "username": "sheldon cooper",
        "id": "65d09ade1d06de00132f7eb1"
    },
    "category": "music",
    "text": "the Big Bang theory",
    "votes": [
        {
            "user": "65d09ade1d06de00132f7eb1",
            "vote": 1
        }
    ],
    "comments": [
        {
            "created": "2024-02-17T11:51:04.898Z",
            "author": {
                "username": "akro",
                "id": "65d09ade1d06de00132f7eb1"
            },
            "body": "Good!",
            "id": "65d09da81d06de00132f7eb4"
        }
    ],
    "created": "2024-02-17T11:51:01.083Z",
    "upvotePercentage": 100,
    "id": "65d09da51d06de00132f7eb3"
}
```


### Добавить комментарий POST /api/post/{POST_ID}
**Принимает: объект JSON post**  
**Возвращает: объект JSON post**  
**Требование: добавить комментарий может только авторизованный пользователь**

Пример возможного запроса (при условии авторизации):
```bash
curl -H 'Content-Type: application/json' -d '{"comment":"Sheldon Cooper"}' -X 'POST' "http://localhost:8080/api/post/id"
```

Пример успешного ответа:
```json
{
    ...
    "comments": [
        {
            "author": {
                "username": "akro",
                "id": "65d09ade1d06de00132f7eb1"
            },
            "body": "Sheldon Cooper",
            "created": "2024-02-17T12:00:03.847Z",
            "id": "65d09fc31d06de00132f7eb6"
        },
        ...
    ],
    ...
}
```

### Удалить комментарий DELETE /api/post/{POST_ID}/{COMMENT_ID}
**Принимает: -**  
**Возвращает: объект JSON post**  
**Требование: удалить комментарий может только авторизованный пользователь**

Пример возможного запроса (при условии авторизации):
```bash
curl -X 'DELETE' "http://localhost:8080/api/post/id/id"
```

Пример успешного ответа:
До:
```json
...
"comments": [
    {
        "created": "2024-02-17T11:55:19.381Z",
        "author": {
            "username": "akro",
            "id": "65d09ade1d06de00132f7eb1"
        },
        "body": "hello",
        "id": "id"
    }
],
...
```
После:
```json
...
"comments": [],
...
```

### Регистрация пользователя POST /api/register
**Принимает: объект JSON user**  
**Возвращает: объект JSON token**

Пример успешного ответа:
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7InVzZXJuYW1lIjoiZGFuZmphZm4iLCJpZCI6IjY1ZDBhOWQ4MWQwNmRlMDAxMzJmN2ViOSJ9LCJpYXQiOjE3MDgxNzM3ODQsImV4cCI6MTcwODc3ODU4NH0.dVE0h951ZLZVpdBaCtSX-Wuy-YXoeF_xSTXxrSHaLbw"
}
```

### Авторизация пользователя POST /api/login
**Принимает: объект JSON user**  
**Возвращает: объект JSON token**

Пример успешного ответа:
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7InVzZXJuYW1lIjoiZGFuZmphZm4iLCJpZCI6IjY1ZDBhOWQ4MWQwNmRlMDAxMzJmN2ViOSJ9LCJpYXQiOjE3MDgxNzM3ODQsImV4cCI6MTcwODc3ODU4NH0.dVE0h951ZLZVpdBaCtSX-Wuy-YXoeF_xSTXxrSHaLbw"
}
```
