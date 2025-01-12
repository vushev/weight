-- Таблица за настройки на калориите
CREATE TABLE IF NOT EXISTS calorie_settings (
    user_id INT PRIMARY KEY,
    gender ENUM('male', 'female') NOT NULL,
    age INT NOT NULL,
    activity_level ENUM('sedentary', 'light', 'moderate', 'active', 'very_active') NOT NULL,
    goal ENUM('maintain', 'lose', 'gain') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Таблица за дневен лог на калориите
CREATE TABLE IF NOT EXISTS daily_calorie_logs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    date DATE NOT NULL,
    food_kcal DECIMAL(10,2) DEFAULT 0,
    activity_kcal DECIMAL(10,2) DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_user_date (user_id, date)
);

-- Таблица за записи на храна
CREATE TABLE IF NOT EXISTS food_entries (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    log_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    calories DECIMAL(10,2) NOT NULL,
    protein DECIMAL(10,2) DEFAULT 0,
    carbs DECIMAL(10,2) DEFAULT 0,
    fat DECIMAL(10,2) DEFAULT 0,
    time TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (log_id) REFERENCES daily_calorie_logs(id) ON DELETE CASCADE
);

-- Таблица за записи на активности
CREATE TABLE IF NOT EXISTS activity_entries (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    log_id INT NOT NULL,
    type VARCHAR(50) NOT NULL,
    duration INT NOT NULL,
    calories DECIMAL(10,2) NOT NULL,
    time TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (log_id) REFERENCES daily_calorie_logs(id) ON DELETE CASCADE
);

-- Индекси за по-бързо търсене
CREATE INDEX idx_daily_logs_user_date ON daily_calorie_logs(user_id, date);
CREATE INDEX idx_food_entries_user_log ON food_entries(user_id, log_id);
CREATE INDEX idx_activity_entries_user_log ON activity_entries(user_id, log_id); 