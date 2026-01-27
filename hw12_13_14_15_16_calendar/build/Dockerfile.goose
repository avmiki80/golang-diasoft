FROM golang:1.20-alpine

# Установка goose v3.15.0 (совместима с Go 1.20)
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.15.0

WORKDIR /migrations

# Команда по умолчанию - применить миграции
CMD ["sh", "-c", "goose postgres \"host=${DB_HOST} port=${DB_PORT} user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_NAME} sslmode=disable\" up"]