FROM golang:1.21-alpine

WORKDIR /app

# Инсталирайте необходимите зависимости
RUN apk add --no-cache build-base

# Копирайте go.mod и go.sum и изтеглете зависимостите
COPY go.mod go.sum ./
RUN go mod download && go mod tidy

# Копирайте целия проект
COPY . .

# Изложете порта
EXPOSE 8080

# Определете командата за стартиране
CMD ["go", "run", "./cmd/main.go"] 