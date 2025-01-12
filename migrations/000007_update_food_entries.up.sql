-- Създаваме таблица за типовете хранения
CREATE TABLE IF NOT EXISTS meal_types (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Добавяме базови типове хранения
-- INSERT INTO meal_types (name) VALUES 
--     ('Закуска'),
--     ('Обяд'),
--     ('Вечеря'),
--     ('Междинно хранене');

-- Модифицираме таблицата food_entries
ALTER TABLE food_entries 
    ADD COLUMN meal_type_id INT,
    ADD COLUMN notes TEXT,
    ADD FOREIGN KEY (meal_type_id) REFERENCES meal_types(id); 

-- Добавяме колона за време във food_entries
ALTER TABLE food_entries ADD COLUMN time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;