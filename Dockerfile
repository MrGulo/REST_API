# Этап 1: Сборка бинарного файла
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o task_api ./cmd/task_api/main.go

FROM alpine:3.20

RUN apk --no-cache add ca-certificates

RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/task_api .
COPY --from=builder /app/migrations ./migrations

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

CMD ["./task_api"]