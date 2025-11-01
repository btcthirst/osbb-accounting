-- Міграція 002: Таблиця квартир
-- Створює таблицю apartments для зберігання інформації про квартири ОСББ

-- Таблиця квартир
CREATE TABLE IF NOT EXISTS apartments (
    -- Унікальний ідентифікатор квартири
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- Номер квартири (унікальний, може містити букви)
    apartment_number TEXT NOT NULL UNIQUE,
    
    -- Поверх
    floor INTEGER NOT NULL CHECK(floor >= 0),
    
    -- Під'їзд (опціонально, може бути NULL)
    entrance INTEGER,
    
    -- Площа квартири в м²
    area REAL NOT NULL CHECK(area > 0),
    
    -- Кількість кімнат
    rooms INTEGER NOT NULL CHECK(rooms >= 0),
    
    -- ПІБ власника
    owner_name TEXT NOT NULL,
    
    -- Контактний телефон власника
    owner_phone TEXT NOT NULL,
    
    -- Email власника (опціонально)
    owner_email TEXT,
    
    -- Кількість зареєстрованих мешканців
    residents_count INTEGER NOT NULL DEFAULT 0 CHECK(residents_count >= 0),
    
    -- Додаткові примітки
    notes TEXT,
    
    -- Прапорець активності (для м'якого видалення)
    is_active BOOLEAN NOT NULL DEFAULT 1,
    
    -- Час створення запису
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Час останнього оновлення
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Індекс для швидкого пошуку за номером квартири
CREATE INDEX IF NOT EXISTS idx_apartments_number ON apartments(apartment_number);

-- Індекс для пошуку за поверхом
CREATE INDEX IF NOT EXISTS idx_apartments_floor ON apartments(floor);

-- Індекс для пошуку за під'їздом
CREATE INDEX IF NOT EXISTS idx_apartments_entrance ON apartments(entrance);

-- Індекс для фільтрації активних квартир
CREATE INDEX IF NOT EXISTS idx_apartments_active ON apartments(is_active);

-- Індекс для пошуку за власником
CREATE INDEX IF NOT EXISTS idx_apartments_owner ON apartments(owner_name);

-- Композитний індекс для поверху та під'їзду
CREATE INDEX IF NOT EXISTS idx_apartments_floor_entrance ON apartments(floor, entrance);

-- Вставка тестових даних (для розробки)
-- Можна видалити після завершення розробки
INSERT INTO apartments (apartment_number, floor, entrance, area, rooms, owner_name, owner_phone, residents_count, notes)
VALUES 
    ('1', 1, 1, 45.5, 2, 'Іванов Іван Іванович', '+380501234567', 2, 'Перша квартира'),
    ('2', 1, 1, 52.3, 2, 'Петренко Петро Петрович', '+380502345678', 3, ''),
    ('3', 1, 1, 68.0, 3, 'Сидоренко Сидір Сидорович', '+380503456789', 4, 'Велика сім`я'),
    ('4', 2, 1, 45.5, 2, 'Коваленко Ольга Іванівна', '+380504567890', 1, ''),
    ('5', 2, 1, 52.3, 2, 'Мельник Марія Петрівна', '+380505678901', 2, ''),
    ('6', 2, 1, 68.0, 3, 'Шевченко Тарас Григорович', '+380506789012', 3, 'Квартира з балконом');

-- Примітка: Тестові дані додані для зручності розробки та тестування UI