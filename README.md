# PR Reviewer Assignment Service

Микросервис для автоматического назначения ревьюверов на Pull Request'ы с управлением командами и пользователями.

## Инструкция по запуску

1. Клонирование репозитория
```
git clone https://github.com/Shish-ai-ai/pr-reviewer-assignment.git
cd pr-reviewer-assignment
```

2. Запуск с помощью Docker Compose
```
docker-compose up --build -d
```

После успешного запуска сервис будет доступен по адресу `http://localhost:8082`.

## API эндпоинты

### 1. Создание команды с участниками
**POST** `http://localhost:8082/team/add`

Тело запроса:
```json
{
    "team_name": "backend",
    "members": [
        {
            "user_id": "u1",
            "username": "Alice",
            "is_active": true
        },
        {
            "user_id": "u2", 
            "username": "Bob",
            "is_active": true
        }
    ]
}
```

Ответ:
```json
{
    "team": {
        "team_name": "backend",
        "members": [
            {
                "user_id": "u1",
                "username": "Alice",
                "is_active": true
            },
            {
                "user_id": "u2",
                "username": "Bob",
                "is_active": true
            }
        ]
    }
}
```

### 2. Получение команды с участниками
**GET** `http://localhost:8082/team/get?team_name=backend`

Ответ:
```json
{
    "team_name": "backend",
    "members": [
        {
            "user_id": "u1",
            "username": "Alice",
            "is_active": true
        },
        {
            "user_id": "u2",
            "username": "Bob",
            "is_active": true
        }
    ]
}
```

### 3. Установка флага активности пользователя
**POST** `http://localhost:8082/users/setIsActive`

Тело запроса:
```json
{
    "user_id": "u2",
    "is_active": false
}
```

Ответ:
```json
{
    "user": {
        "user_id": "u2",
        "username": "Bob",
        "team_name": "backend",
        "is_active": false
    }
}
```

### 4. Получение PR'ов, где пользователь назначен ревьювером
**GET** `http://localhost:8082/users/getReview?user_id=u2`

Ответ:
```json
{
    "user_id": "u2",
    "pull_requests": [
        {
            "pull_request_id": "pr-1001",
            "pull_request_name": "Add search",
            "author_id": "u1",
            "status": "OPEN"
        }
    ]
}
```

### 5. Создание PR и автоматическое назначение ревьюверов
**POST** `http://localhost:8082/pullRequest/create`

Тело запроса:
```json
{
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add search",
    "author_id": "u1"
}
```

Ответ:
```json
{
    "pr": {
        "pull_request_id": "pr-1001",
        "pull_request_name": "Add search",
        "author_id": "u1",
        "status": "OPEN",
        "assigned_reviewers": ["u2", "u3"],
        "createdAt": "2025-11-22T14:30:34.278941652Z"
    }
}
```

### 6. Мерж PR (идемпотентная операция)
**POST** `http://localhost:8082/pullRequest/merge`

Тело запроса:
```json
{
    "pull_request_id": "pr-1001"
}
```

Ответ:
```json
{
    "pr": {
        "pull_request_id": "pr-1001",
        "pull_request_name": "Add search",
        "author_id": "u1",
        "status": "MERGED",
        "assigned_reviewers": ["u2", "u3"],
        "mergedAt": "2025-11-22T15:45:56.7320526Z"
    }
}
```

### 7. Переназначение ревьювера
**POST** `http://localhost:8082/pullRequest/reassign`

Тело запроса:
```json
{
    "pull_request_id": "pr-1001",
    "old_reviewer_id": "u2"
}
```

Ответ:
```json
{
    "pr": {
        "pull_request_id": "pr-1001",
        "pull_request_name": "Add search",
        "author_id": "u1",
        "status": "OPEN",
        "assigned_reviewers": ["u5", "u3"]
    },
    "replaced_by": "u5"
}
```

### 8. Статистика по ревьюверам
**GET** `http://localhost:8082/stats/reviewers`

Ответ:
```json
{
    "reviewer_stats": [
        {
            "user_id": "u2",
            "count": 3
        },
        {
            "user_id": "u3",
            "count": 2
        }
    ]
}
```

## Коды ошибок

- `TEAM_EXISTS` - команда уже существует
- `PR_EXISTS` - PR уже существует
- `PR_MERGED` - нельзя изменить мерженный PR
- `NOT_ASSIGNED` - пользователь не назначен ревьювером
- `NO_CANDIDATE` - нет доступных кандидатов для замены
- `NOT_FOUND` - ресурс не найден