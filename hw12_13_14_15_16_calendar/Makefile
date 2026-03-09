# Бинарные файлы
BIN_CALENDAR := "./bin/calendar"
BIN_SCHEDULER := "./bin/scheduler"
BIN_STORER := "./bin/storer"
BIN_DLQ_MONITOR := "./bin/dlq-monitor"

# Docker образы
DOCKER_IMG_CALENDAR := "calendar:develop"
DOCKER_IMG_SCHEDULER := "scheduler:develop"
DOCKER_IMG_STORER := "storer:develop"
DOCKER_IMG_DLQ_MONITOR := "dlq-monitor:develop"

# Версионирование
GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

# ============================================================================
# Build targets
# ============================================================================

build: build-calendar build-scheduler build-storer build-dlq-monitor

build-calendar:
	go build -v -o $(BIN_CALENDAR) -ldflags "$(LDFLAGS)" ./cmd/calendar

build-scheduler:
	go build -v -o $(BIN_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/scheduler

build-storer:
	go build -v -o $(BIN_STORER) -ldflags "$(LDFLAGS)" ./cmd/storer

build-dlq-monitor:
	go build -v -o $(BIN_DLQ_MONITOR) -ldflags "$(LDFLAGS)" ./cmd/dlq-monitor

# ============================================================================
# Run targets
# ============================================================================

run-calendar: build-calendar
	$(BIN_CALENDAR) -config ./configs/calendar/config.yaml

run-scheduler: build-scheduler
	$(BIN_SCHEDULER) -config ./configs/scheduler/config.yaml

run-storer: build-storer
	$(BIN_STORER) -config ./configs/storer/config.yaml

run-dlq-monitor: build-dlq-monitor
	$(BIN_DLQ_MONITOR) -config ./configs/dlq-monitor/config.yaml

# Для обратной совместимости
run: run-calendar

# ============================================================================
# Version targets
# ============================================================================

version-calendar: build-calendar
	$(BIN_CALENDAR) version

version-scheduler: build-scheduler
	$(BIN_SCHEDULER) version

version-storer: build-storer
	$(BIN_STORER) version

version-dlq-monitor: build-dlq-monitor
	$(BIN_DLQ_MONITOR) version

version: version-calendar version-scheduler version-storer version-dlq-monitor

# ============================================================================
# Docker targets
# ============================================================================

build-img: build-img-calendar build-img-scheduler build-img-storer build-img-dlq-monitor

build-img-calendar:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		--build-arg=SERVICE=calendar \
		-t $(DOCKER_IMG_CALENDAR) \
		-f build/Dockerfile .

build-img-scheduler:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		--build-arg=SERVICE=scheduler \
		-t $(DOCKER_IMG_SCHEDULER) \
		-f build/Dockerfile .

build-img-storer:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		--build-arg=SERVICE=storer \
		-t $(DOCKER_IMG_STORER) \
		-f build/Dockerfile .

build-img-dlq-monitor:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		--build-arg=SERVICE=dlq-monitor \
		-t $(DOCKER_IMG_DLQ_MONITOR) \
		-f build/Dockerfile .

run-img-calendar: build-img-calendar
	docker run $(DOCKER_IMG_CALENDAR)

run-img-scheduler: build-img-scheduler
	docker run $(DOCKER_IMG_SCHEDULER)

run-img-storer: build-img-storer
	docker run $(DOCKER_IMG_STORER)

run-img-dlq-monitor: build-img-dlq-monitor
	docker run $(DOCKER_IMG_DLQ_MONITOR)

# Для обратной совместимости
run-img: run-img-calendar

# ============================================================================
# Testing
# ============================================================================

test:
	#go test -race ./internal/repositories/memory/... ./pkg/...
	go test -race ./internal/...

# Unit-тесты (без БД, только memory storage)
test-unit:
	@echo "Running unit tests (memory storage)..."
	go test -v -race ./internal/repositories/memory/...

# Интеграционные тесты с testcontainers (автоматически запускает PostgreSQL)
test-integration:
	@echo "Running integration tests with testcontainers..."
	go test -v -race -tags=integration ./internal/repositories/db/...
	go test -v -race -tags=integration ./internal/services/...

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
	go test -race -tags=integration -coverprofile=coverage-integration.out ./internal/repositories/db/... ./internal/services/...
	go tool cover -html=coverage-integration.out -o coverage-integration.html
	@echo "Integration coverage report generated: coverage-integration.html"

# ============================================================================
# Code Quality
# ============================================================================

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.64.6

lint: install-lint-deps
	golangci-lint run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

# ============================================================================
# Database
# ============================================================================

migrate:
	@echo "Applying migrations..."
	@which goose > /dev/null || (echo "goose not installed. Run: go install github.com/pressly/goose/v3/cmd/goose@latest" && exit 1)
	cd migrations && goose postgres "postgres://postgres:postgres@localhost:5432/calendar?sslmode=disable" up

migrate-down:
	@echo "Rolling back last migration..."
	@which goose > /dev/null || (echo "goose not installed. Run: go install github.com/pressly/goose/v3/cmd/goose@latest" && exit 1)
	cd migrations && goose postgres "postgres://postgres:postgres@localhost:5432/calendar?sslmode=disable" down

migrate-status:
	@echo "Migration status..."
	@which goose > /dev/null || (echo "goose not installed. Run: go install github.com/pressly/goose/v3/cmd/goose@latest" && exit 1)
	cd migrations && goose postgres "postgres://postgres:postgres@localhost:5432/calendar?sslmode=disable" status

# ============================================================================
# Docker Compose
# ============================================================================
up:
	cd deployments && docker-compose up -d

down:
	cd deployments && docker-compose down

down-volumes:
	cd deployments && docker-compose down -v

logs:
	cd deployments && docker-compose logs -f

logs-calendar:
	cd deployments && docker-compose logs -f calendar

logs-scheduler:
	cd deployments && docker-compose logs -f scheduler

logs-storer:
	cd deployments && docker-compose logs -f storer

logs-dlq-monitor:
	cd deployments && docker-compose logs -f dlq-monitor

restart:
	cd deployments && docker-compose restart

restart-calendar:
	cd deployments && docker-compose restart calendar

restart-scheduler:
	cd deployments && docker-compose restart scheduler

restart-storer:
	cd deployments && docker-compose restart storer

restart-dlq-monitor:
	cd deployments && docker-compose restart dlq-monitor

# Пересоздать БД с миграциями
recreate-db: down-volumes up
	@echo "Database recreated with migrations applied"

# Применить миграции в docker-compose
docker-migrate:
	cd deployments && docker-compose up goose-migrate

# ============================================================================
# Utilities
# ============================================================================

clean:
	@echo "Cleaning build artifacts..."
	rm -rf ./bin/* coverage*.out coverage*.html
	@echo "Clean complete"

deps:
	go mod download
	go mod tidy

# ============================================================================
# Help
# ============================================================================

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

.PHONY: build build-calendar build-scheduler build-storer build-dlq-monitor \
	run run-calendar run-scheduler run-storer run-dlq-monitor \
	version version-calendar version-scheduler version-storer version-dlq-monitor \
	build-img build-img-calendar build-img-scheduler build-img-storer build-img-dlq-monitor \
	run-img run-img-calendar run-img-scheduler run-img-storer run-img-dlq-monitor \
	test test-unit test-integration test-all test-coverage test-coverage-integration \
	lint fmt vet install-lint-deps \
	migrate migrate-down migrate-status docker-migrate \
	up down down-volumes logs logs-calendar logs-scheduler logs-storer logs-dlq-monitor \
	restart restart-calendar restart-scheduler restart-storer restart-dlq-monitor \
	recreate-db clean deps help
