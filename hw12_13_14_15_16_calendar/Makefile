BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

run: build
	$(BIN) -config ./configs/config.toml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

test:
	go test -race ./internal/repositories/memory/... ./pkg/...

# Unit-тесты (без БД, только memory storage)
test-unit:
	@echo "Running unit tests (memory storage)..."
	go test -v -race ./internal/repositories/memory/...

# Интеграционные тесты с testcontainers (автоматически запускает PostgreSQL)
test-integration:
	@echo "Running integration tests with testcontainers..."
	go test -v -race -tags=integration ./internal/repositories/db/...

# Все тесты
test-all: test-unit test-integration

# Тесты с покрытием
test-coverage:
	@echo "Running tests with coverage..."
	go test -race -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Тесты с покрытием для integration
test-coverage-integration:
	@echo "Running integration tests with coverage..."
	go test -race -tags=integration -coverprofile=coverage-integration.out ./internal/repositories/db/...
	go tool cover -html=coverage-integration.out -o coverage-integration.html
	@echo "Integration coverage report generated: coverage-integration.html"

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2

migrate:
	@echo "Applying migrations..."
	@which goose > /dev/null || (echo "goose not installed. Run: go install github.com/pressly/goose/v3/cmd/goose@latest" && exit 1)
	cd migrations && goose postgres "postgres://postgres:postgres@localhost:5432/calendar?sslmode=disable" up

# Docker Compose команды
up:
	cd deployments && docker-compose up -d

down:
	cd deployments && docker-compose down

down-volumes:
	cd deployments && docker-compose down -v

logs:
	cd deployments && docker-compose logs -f

restart:
	cd deployments && docker-compose restart

# Пересоздать БД с миграциями
recreate-db: down-volumes up
	@echo "Database recreated with migrations applied"

# Применить миграции в docker-compose
docker-migrate:
	cd deployments && docker-compose up goose-migrate

lint: install-lint-deps
	golangci-lint run ./...

# Очистка артефактов
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN) coverage*.out coverage*.html
	@echo "Clean complete"

# Помощь
help:
	@echo "Available targets:"
	@echo "  build                - Компилировать бинарный файл"
	@echo "  run                  - Собрать и запустить сервис"
	@echo "  version              - Показать версию"
	@echo ""
	@echo "Testing:"
	@echo "  test                 - Запустить все тесты с -race"
	@echo "  test-unit            - Запустить unit-тесты (memory storage)"
	@echo "  test-integration     - Запустить integration-тесты (testcontainers)"
	@echo "  test-all             - Запустить все тесты (unit + integration)"
	@echo "  test-coverage        - Тесты с покрытием"
	@echo "  test-coverage-integration - Integration тесты с покрытием"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint                 - Запустить golangci-lint"
	@echo "  install-lint-deps    - Установить golangci-lint"
	@echo ""
	@echo "Database:"
	@echo "  migrate              - Применить миграции"
	@echo "  migrate-down         - Откатить миграции"
	@echo "  migrate-status       - Статус миграций"
	@echo ""
	@echo "Docker:"
	@echo "  build-img            - Собрать Docker образ"
	@echo "  run-img              - Запустить Docker контейнер"
	@echo "  up                   - Запустить docker-compose"
	@echo "  down                 - Остановить docker-compose"
	@echo "  down-volumes         - Остановить docker-compose с удалением volumes"
	@echo "  logs                 - Показать логи docker-compose"
	@echo "  restart              - Перезапустить docker-compose"
	@echo "  recreate-db          - Пересоздать БД с миграциями"
	@echo "  docker-migrate       - Применить миграции в docker-compose"
	@echo ""
	@echo "Other:"
	@echo "  clean                - Очистить артефакты сборки"
	@echo "  help                 - Показать эту помощь"

.PHONY: build run build-img run-img version test test-unit test-integration test-all test-coverage test-coverage-integration lint migrate migrate-down migrate-status up down down-volumes logs restart recreate-db docker-migrate install-lint-deps clean help
