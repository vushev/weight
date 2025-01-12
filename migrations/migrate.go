package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "time"

    _ "github.com/go-sql-driver/mysql"
    "github.com/joho/godotenv"
)

func main() {
    // Зареждане на .env файл
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }

    // Създаване на връзка с базата данни
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_HOST"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_NAME"),
        os.Getenv("DB_CHARSET"))

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Проверка на връзката
    if err = db.Ping(); err != nil {
        log.Fatal("Could not connect to database:", err)
    }

    log.Println("Successfully connected to database")

    // Създаване на таблица за миграции ако не съществува
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS migrations (
            id INT AUTO_INCREMENT PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        log.Fatal("Error creating migrations table:", err)
    }

    // Списък с миграции
    migrations := []struct {
        name string
        up   string
    }{
        {
            name: "create_users_table",
            up: `CREATE TABLE IF NOT EXISTS users (
                id INT AUTO_INCREMENT PRIMARY KEY,
                username VARCHAR(255) NOT NULL UNIQUE,
                password VARCHAR(255) NOT NULL,
                first_name VARCHAR(255),
                last_name VARCHAR(255),
                age INT,
                height FLOAT,
                gender VARCHAR(10),
                email VARCHAR(255),
                target_weight FLOAT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
            )`,
        },
        {
            name: "create_weight_records_table",
            up: `CREATE TABLE IF NOT EXISTS weight_records (
                id INT AUTO_INCREMENT PRIMARY KEY,
                user_id INT NOT NULL,
                weight FLOAT NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (user_id) REFERENCES users(id)
            )`,
        },
        {
            name: "create_user_settings_table",
            up: `CREATE TABLE IF NOT EXISTS user_settings (
                id INT AUTO_INCREMENT PRIMARY KEY,
                user_id INT NOT NULL UNIQUE,
                is_visible BOOLEAN DEFAULT true,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                FOREIGN KEY (user_id) REFERENCES users(id)
            )`,
        },
        {
            name: "create_friendships_table",
            up: `CREATE TABLE IF NOT EXISTS friendships (
                id INT AUTO_INCREMENT PRIMARY KEY,
                requester_id INT NOT NULL,
                addressee_id INT NOT NULL,
                status ENUM('pending', 'accepted', 'rejected') DEFAULT 'pending',
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                FOREIGN KEY (requester_id) REFERENCES users(id),
                FOREIGN KEY (addressee_id) REFERENCES users(id)
            )`,
        },
        {
            name: "create_challenges_table",
            up: `CREATE TABLE IF NOT EXISTS challenges (
                id INT AUTO_INCREMENT PRIMARY KEY,
                creator_id INT NOT NULL,
                opponent_id INT NOT NULL,
                start_date DATE NOT NULL,
                end_date DATE NOT NULL,
                status ENUM('pending', 'active', 'completed', 'rejected') DEFAULT 'pending',
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                FOREIGN KEY (creator_id) REFERENCES users(id),
                FOREIGN KEY (opponent_id) REFERENCES users(id)
            )`,
        },
        {
            name: "create_challenge_results_table",
            up: `CREATE TABLE IF NOT EXISTS challenge_results (
                id INT AUTO_INCREMENT PRIMARY KEY,
                challenge_id INT NOT NULL,
                user_id INT NOT NULL,
                initial_weight FLOAT NOT NULL,
                final_weight FLOAT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                FOREIGN KEY (challenge_id) REFERENCES challenges(id),
                FOREIGN KEY (user_id) REFERENCES users(id)
            )`,
        },
        {
            name: "create_calorie_settings_table",
            up: `CREATE TABLE IF NOT EXISTS calorie_settings (
                id INT AUTO_INCREMENT PRIMARY KEY,
                user_id INT NOT NULL UNIQUE,
                gender VARCHAR(10) NOT NULL,
                age INT NOT NULL,
                activity_level VARCHAR(20) NOT NULL,
                goal VARCHAR(20) NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                FOREIGN KEY (user_id) REFERENCES users(id)
            )`,
        },
        {
            name: "create_daily_calorie_logs_table",
            up: `CREATE TABLE IF NOT EXISTS daily_calorie_logs (
                id INT AUTO_INCREMENT PRIMARY KEY,
                user_id INT NOT NULL,
                date DATE NOT NULL,
                food_kcal FLOAT DEFAULT 0,
                activity_kcal FLOAT DEFAULT 0,
                notes TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                FOREIGN KEY (user_id) REFERENCES users(id),
                UNIQUE KEY unique_user_date (user_id, date)
            )`,
        },
        {
            name: "create_food_entries_table",
            up: `CREATE TABLE IF NOT EXISTS food_entries (
                id INT AUTO_INCREMENT PRIMARY KEY,
                user_id INT NOT NULL,
                log_id INT NOT NULL,
                name VARCHAR(255) NOT NULL,
                calories FLOAT NOT NULL,
                protein FLOAT,
                carbs FLOAT,
                fat FLOAT,
                time TIMESTAMP NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (user_id) REFERENCES users(id),
                FOREIGN KEY (log_id) REFERENCES daily_calorie_logs(id)
            )`,
        },
        {
            name: "create_activity_entries_table",
            up: `CREATE TABLE IF NOT EXISTS activity_entries (
                id INT AUTO_INCREMENT PRIMARY KEY,
                user_id INT NOT NULL,
                log_id INT NOT NULL,
                type VARCHAR(255) NOT NULL,
                duration INT NOT NULL,
                calories FLOAT NOT NULL,
                time TIMESTAMP NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (user_id) REFERENCES users(id),
                FOREIGN KEY (log_id) REFERENCES daily_calorie_logs(id)
            )`,
        },
    }

    // Изпълнение на миграциите
    for _, migration := range migrations {
        // Проверка дали миграцията вече е изпълнена
        var exists bool
        err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM migrations WHERE name = ?)", migration.name).Scan(&exists)
        if err != nil {
            log.Fatal("Error checking migration:", err)
        }

        if !exists {
            // Изпълнение на миграцията
            log.Printf("Applying migration: %s", migration.name)
            
            tx, err := db.Begin()
            if err != nil {
                log.Fatal("Error starting transaction:", err)
            }

            // Изпълнение на SQL заявките
            if _, err := tx.Exec(migration.up); err != nil {
                tx.Rollback()
                log.Fatal("Error executing migration:", err)
            }

            // Записване на миграцията
            if _, err := tx.Exec("INSERT INTO migrations (name) VALUES (?)", migration.name); err != nil {
                tx.Rollback()
                log.Fatal("Error recording migration:", err)
            }

            if err := tx.Commit(); err != nil {
                log.Fatal("Error committing transaction:", err)
            }

            log.Printf("Successfully applied migration: %s", migration.name)
        } else {
            log.Printf("Skipping migration %s (already applied)", migration.name)
        }
    }

    log.Println("All migrations completed successfully")
} 