# Reviewer Appointment Service

Сервис для автоматического назначения ревьюверов на Pull Request'ы.

## Описание

Микросервис, который автоматически назначает ревьюверов на PR из команды автора, позволяет выполнять переназначение ревьюверов и получать список PR'ов, назначенных конкретному пользователю, а также управлять командами и активностью пользователей.

## Требования

- Go 1.24+
- PostgreSQL 15+
- Docker и Docker Compose (опционально)

## Быстрый старт

### С использованием Docker Compose (рекомендуется)

```bash
docker-compose up -d
```

Сервис будет доступен на `http://localhost:8081`

PostgreSQL и сервис запустятся автоматически, миграции применятся при старте.

### Локальный запуск

1. Убедитесь, что PostgreSQL запущен и доступен

2. Создайте базу данных (если еще не создана):
```sql
CREATE DATABASE reviewer_appointment;
```

3. Настройте подключение через переменные окружения или `config/config.yaml`:
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=reviewer_appointment
```

4. Запустите сервис:
```bash
make run
```

или

```bash
go run ./cmd/reviewer-appointment-service
```

**Примечание:** Если PostgreSQL не запущен, вы увидите понятное сообщение об ошибке с инструкциями.

## API Endpoints

### Teams
- `POST /team/add` - Создать команду с участниками
- `GET /team/get?team_name=...` - Получить команду с участниками

### Users
- `POST /users/setIsActive` - Установить флаг активности пользователя
- `GET /users/getReview?user_id=...` - Получить PR'ы, где пользователь назначен ревьювером

### Pull Requests
- `POST /pullRequest/create` - Создать PR и автоматически назначить до 2 ревьюверов
- `POST /pullRequest/merge` - Пометить PR как MERGED (идемпотентная операция)
- `POST /pullRequest/reassign` - Переназначить ревьювера на другого из его команды

### Stats
- `GET /stats` - Получить статистику по сервису

### Health
- `GET /health` - Проверка работоспособности сервиса

## Архитектура

Проект следует чистой архитектуре:

```
cmd/
  reviewer-appointment-service/
    main.go - точка входа
internal/
  config/ - конфигурация
  models/domain/ - доменные модели
  services/ - бизнес-логика
  handlers/ - HTTP handlers
  storage/ - слой работы с БД
    postgresql/ - реализация для PostgreSQL
migrations/ - SQL миграции
config/ - конфигурационные файлы
```

## Тестирование

```bash
make test
```

Для запуска тестов хранилищ требуется доступ к PostgreSQL. Настройте переменные окружения:

```bash
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5432
export TEST_DB_USER=postgres
export TEST_DB_PASSWORD=postgres
export TEST_DB_NAME=reviewer_appointment_test
```

## Сборка

```bash
make build
```

Бинарный файл будет создан в `bin/reviewer-appointment-service`

## Docker

### Сборка образа
```bash
docker build -t reviewer-appointment-service .
```

### Запуск с Docker Compose
```bash
make docker-up
```

### Остановка
```bash
make docker-down
```

## Конфигурация

Конфигурация находится в `config/config.yaml`. Поддерживаются переменные окружения для переопределения значений.

## Вопросы и решения

### Вопросы, которые возникли при разработке:

1. **Использование Gin вместо net/http**: Выбран Gin для упрощения работы с HTTP и лучшей поддержки middleware.

2. **Автоматическое применение миграций**: Миграции применяются автоматически при старте приложения для упрощения развертывания.

3. **Обработка ошибок**: Используется кастомный тип ошибок `AppError` для единообразной обработки ошибок в handlers.

4. **Статистика**: Реализован базовый эндпоинт статистики с возможностью расширения в будущем.

