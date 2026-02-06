.PHONY: run test docker-up docker-down migrate-up migrate-down

# Запуск приложения локально
run:
	go run ./cmd/api

# Запуск тестов
test:
	go test ./tests/ -v -count=1

# Поднять Docker (PostgreSQL + приложение)
docker-up:
	docker-compose up -d

# Остановить Docker
docker-down:
	docker-compose down

# Только PostgreSQL
postgres-up:
	docker-compose up -d postgres

postgres-down:
	docker-compose down postgres
