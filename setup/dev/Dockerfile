FROM golang:1.21-alpine

WORKDIR /app

# Инсталирайте необходимите зависимости
RUN apk add --no-cache build-base

# Създаваме директория за логове и задаваме правилните права
RUN mkdir -p /app/logs && chmod 755 /app/logs

# Копирайте go.mod и go.sum от главната директория
COPY ../../go.mod ../../go.sum ./
RUN go mod download && go mod tidy

# Копирайте целия проект
COPY ../../. .

# Изложете порта
EXPOSE 8080

# Определете командата за стартиране
CMD ["go", "run", "./cmd/main.go"]