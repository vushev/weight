FROM golang:1.21-alpine

WORKDIR /app

# Инсталиране на build-base за компилация
RUN apk add --no-cache build-base

# Копиране на go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download && go mod tidy

# Копиране на сорс кода
COPY . .

# Обновяване на зависимостите и билдване
RUN go mod tidy && CGO_ENABLED=1 GOOS=linux go build -o main .

# Експозване на порт
EXPOSE 8080

# Стартиране на приложението
CMD ["./main"] 