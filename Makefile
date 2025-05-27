.PHONY: migrate-create migrate-up migrate-down

DB_URL = postgres://user:pass@localhost:5432/db_name?sslmode=disable

# Создать новую миграцию
migrate-create:
	@read -p "Введите название миграции: " name; \
	goose -dir migrations create $${name} sql

# Применить миграции
migrate-up:
	goose -dir migrations postgres "$(DB_URL)" up

# Откатить последнюю миграцию
migrate-down:
	goose -dir migrations postgres "$(DB_URL)" down

# Показать статус
migrate-status:
	goose -dir migrations postgres "$(DB_URL)" status

# Запустить приложение и PostgreSQL через Docker Compose
run-docker:
	docker compose --file docker/compose.yml --env-file docker/.env up -d

# Остановить и удалить контейнеры
stop-docker:
	docker compose --file docker/compose.yml --env-file docker/.env down

recreate-docker:
	docker compose --file docker/compose.yml --env-file docker/.env down -v
	docker compose --file docker/compose.yml --env-file docker/.env up --build -d

.PHONY: swagger
swagger:
	@echo "Generating Swagger docs..."
	@swag init -g ./cmd/app/main.go -o ./docs
	@echo "Swagger docs generated in ./docs"

# Команды для работы с seeder
.PHONY: seed-db
seed-db: run-docker
	@echo "Starting database seeding..."
	docker compose --file docker/compose.yml --profile seed up -d db-seeder

.PHONY: load
load: run-docker
	@echo "Starting loading test..."
	/bin/bash ./tests/load/load.sh