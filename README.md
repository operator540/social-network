# GoNetwork

**Социальная сеть на Go** - учебный проект с профессиональной архитектурой.

```
Go + chi v5 + PostgreSQL 16 + JWT + Docker
```

---

## Возможности

- Регистрация и авторизация (access + refresh JWT-токены)
- Создание и удаление постов
- Лайки (с подсчётом в ленте)
- Комментарии к постам
- Подписки на пользователей
- Лента подписок
- Профили с аватарками
- SPA-фронтенд с тёмной темой

## Архитектура

```
HTTP-запрос → Handler → Service → Repository → PostgreSQL
```

| Слой | Ответственность |
|------|----------------|
| **Handler** | Парсинг HTTP, валидация, JSON-ответы |
| **Service** | Бизнес-логика, хеширование, JWT |
| **Repository** | SQL-запросы через интерфейсы |
| **Model** | Чистые структуры данных |

DI через конструкторы в `main.go` - без фреймворков.

## Стек

| Технология | Назначение |
|-----------|-----------|
| [chi v5](https://github.com/go-chi/chi) | HTTP-роутер |
| [PostgreSQL 16](https://www.postgresql.org/) | База данных |
| [lib/pq](https://github.com/lib/pq) | PostgreSQL-драйвер |
| [golang-jwt](https://github.com/golang-jwt/jwt) | JWT-токены |
| [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) | Хеширование паролей |
| [Docker Compose](https://docs.docker.com/compose/) | Оркестрация |

## API — 22 эндпоинта

### Публичные

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/v1/health` | Healthcheck |
| `POST` | `/v1/auth/register` | Регистрация |
| `POST` | `/v1/auth/login` | Логин |
| `POST` | `/v1/auth/refresh` | Обновить токен |
| `GET` | `/v1/feed` | Глобальная лента |
| `GET` | `/v1/users/{id}` | Профиль пользователя |
| `GET` | `/v1/users/{id}/followers` | Подписчики |
| `GET` | `/v1/users/{id}/following` | Подписки |
| `GET` | `/v1/posts/{id}/comments` | Комментарии |

### Защищённые (JWT)

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/v1/auth/logout` | Выход |
| `GET` | `/v1/users/me` | Свой профиль |
| `PUT` | `/v1/users/me` | Обновить bio |
| `POST` | `/v1/users/me/avatar` | Загрузить аватарку |
| `GET` | `/v1/feed/following` | Лента подписок |
| `POST` | `/v1/posts` | Создать пост |
| `DELETE` | `/v1/posts/{id}` | Удалить пост |
| `POST` | `/v1/posts/{id}/comments` | Комментарий |
| `POST` | `/v1/posts/{id}/like` | Лайк |
| `DELETE` | `/v1/posts/{id}/like` | Убрать лайк |
| `POST` | `/v1/users/{id}/follow` | Подписаться |
| `DELETE` | `/v1/users/{id}/follow` | Отписаться |

## База данных — 6 таблиц

```
users ─┬─── posts ──── likes
       │      └─────── comments
       └──── follows
       └──── refresh_tokens
```

## Быстрый старт

```bash
# 1. Клонировать
git clone https://github.com/operator540/social-network.git
cd social-network

# 2. Поднять PostgreSQL
docker compose up -d postgres

# 3. Запустить
go run ./cmd/api

# 4. Открыть
# http://localhost:8080
```

## Docker (полный запуск)

```bash
docker compose up -d
# http://localhost:8080
```

## Тесты

```bash
# Создать тестовую БД (один раз)
docker exec social_postgres psql -U postgres -c "CREATE DATABASE social_network_test;"

# Запустить 30 интеграционных тестов
make test
```

## Структура проекта

```
social-network/
├── cmd/api/main.go              # Точка входа, DI, graceful shutdown
├── internal/
│   ├── config/config.go         # ENV-конфигурация
│   ├── database/postgres.go     # Подключение + миграции
│   ├── model/                   # Структуры данных
│   ├── repository/              # SQL-запросы (интерфейсы)
│   ├── service/                 # Бизнес-логика
│   └── handler/                 # HTTP-хендлеры, роутер, middleware
├── migrations/                  # 6 таблиц (up + down)
├── tests/                       # 30 интеграционных тестов
├── web/index.html               # SPA-фронтенд
├── Dockerfile                   # Multi-stage сборка
└── docker-compose.yml           # PostgreSQL + приложение
```

## Лицензия

MIT
