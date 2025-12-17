# Используем минимальный образ с Go
FROM golang:1.23-alpine

# Рабочая директория
WORKDIR /app

# Копируем код
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 go build -o /bot cmd/main.go

# Убираем лишнее
RUN apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Запуск бота
CMD ["/bot"]