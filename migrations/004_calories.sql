-- +migrate Up
CREATE TABLE calorie_settings (
    user_id INT PRIMARY KEY,
    gender ENUM('male', 'female') NOT NULL,
    age INT NOT NULL,
    activity_level ENUM('sedentary', 'light', 'moderate', 'active', 'very_active') NOT NULL,
    goal ENUM('maintain', 'lose', 'gain') NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE calorie_intake (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    date DATE NOT NULL,
    calories INT NOT NULL,
    type ENUM('food', 'exercise') NOT NULL,
    meal ENUM('breakfast', 'lunch', 'dinner', 'snack') NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_date (user_id, date)
);
