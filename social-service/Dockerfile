# Step 1: Modules caching
FROM golang:1.24.1-alpine AS modules
WORKDIR /modules
COPY go.mod go.sum ./
RUN go mod download

# Step 2: Builder
FROM golang:1.24.1-alpine AS builder
WORKDIR /app
COPY --from=modules /go/pkg /go/pkg
COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    export PATH=$PATH:$(go env GOPATH)/bin

# Generate Swagger docs
RUN $(go env GOPATH)/bin/swag init -g ./cmd/app/main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -tags migrate -o /app/app ./cmd/app

# Step 3: Final Image
FROM alpine:latest
WORKDIR /app

# Копируем конфигурацию и бинарник
COPY --from=builder /app/config /app/config
COPY --from=builder /app/config /config
COPY --from=builder /app/app /app/app
COPY --from=builder /app/pkg/casbin /app/pkg/casbin
COPY --from=builder /app/pkg/postgres /app/pkg/postgres
COPY --from=builder /app/docs /app/docs

# Копируем .env (если не используется volumes в docker-compose)
COPY ../.env /app/.env

# Устанавливаем зависимости (например, клиент PostgreSQL)
RUN apk add --no-cache postgresql-client

# Запуск приложения
CMD ["/app/app"]
