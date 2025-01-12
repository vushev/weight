#!/bin/sh

# Зареждане на .env файл
if [ -f ../.env ]; then
    export $(cat ../.env | grep -v '^#' | xargs)
else
    echo "Error: .env file not found"
    exit 1
fi

# Проверка за базовите MySQL променливи
if [ -z "$DB_HOST" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ] || [ -z "$DB_NAME" ]; then
    echo "Error: Missing required MySQL environment variables"
    echo "Please make sure your .env file contains: DB_HOST, DB_USER, DB_PASSWORD, DB_NAME"
    exit 1
fi

# Функция за изпълнение на миграции локално (в Docker)
run_local_migrations() {
    echo "Running migrations in Docker container..."
    
    # Изпълняваме миграциите директно в контейнера
    docker-compose -f ../docker-compose.dev.yml exec -T db mysql -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < tmp.sql

    # docker-compose -f ../docker-compose.dev.yml exec -T db mysqldump --default-character-set=utf8mb4 -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < tmp.sql

    return $?
}

# Функция за изпълнение на миграции отдалечено
run_remote_migrations() {
    if [ -z "$SSH_HOST" ] || [ -z "$SSH_USER" ]; then
        echo "Error: Missing SSH environment variables"
        echo "Please make sure your .env file contains: SSH_HOST, SSH_USER"
        exit 1
    fi

    echo "Copying migrations file to remote server..."
    scp migrations.sql "$SSH_USER@$SSH_HOST:/tmp/migrations.sql"

    if [ $? -ne 0 ]; then
        echo "Error: Failed to copy migrations file"
        exit 1
    fi

    echo "Running migrations on remote server..."
    ssh "$SSH_USER@$SSH_HOST" "mysql -h\"$DB_HOST\" -P\"$DB_PORT\" -u\"$DB_USER\" -p\"$DB_PASSWORD\" \"$DB_NAME\" < /tmp/migrations.sql && rm /tmp/migrations.sql"
    return $?
}

# Проверка на средата и изпълнение на съответните миграции
if [ "$APP_ENV" = "production" ]; then
    echo "Running in production environment..."
    run_remote_migrations
else
    echo "Running in development environment..."
    run_local_migrations
fi

# Проверка за грешки
if [ $? -eq 0 ]; then
    echo "Migrations completed successfully"
else
    echo "Error running migrations"
    exit 1
fi 