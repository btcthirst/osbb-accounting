-- Міграція 003: Таблиця платежів та нарахувань
-- Створює таблицю payments для обліку фінансових операцій

-- Таблиця платежів
CREATE TABLE IF NOT EXISTS payments (
    -- Унікальний ідентифікатор платежу
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- ID квартири (зовнішній ключ)
    apartment_id INTEGER NOT NULL,
    
    -- Тип платежу: incoming (платіж) або charge (нарахування)
    type TEXT NOT NULL CHECK(type IN ('incoming', 'charge')),
    
    -- Категорія платежу
    category TEXT NOT NULL CHECK(category IN (
        'maintenance', 'repairs', 'utilities', 'heating',
        'water', 'electricity', 'gas', 'other'
    )),
    
    -- Сума платежу (завжди додатня)
    amount REAL NOT NULL CHECK(amount > 0),
    
    -- Опис платежу
    description TEXT NOT NULL,
    
    -- Дата здійснення платежу або нарахування
    payment_date DATE NOT NULL,
    
    -- Період (формат YYYY-MM, наприклад "2025-01")
    period TEXT NOT NULL,
    
    -- ID користувача, який створив запис
    created_by INTEGER NOT NULL,
    
    -- Додаткові примітки
    notes TEXT,
    
    -- Час створення запису
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Час останнього оновлення
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Зовнішні ключі
    FOREIGN KEY (apartment_id) REFERENCES apartments(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id)
);

-- Індекс для швидкого пошуку по квартирі
CREATE INDEX IF NOT EXISTS idx_payments_apartment ON payments(apartment_id);

-- Індекс для пошуку по типу платежу
CREATE INDEX IF NOT EXISTS idx_payments_type ON payments(type);

-- Індекс для пошуку по категорії
CREATE INDEX IF NOT EXISTS idx_payments_category ON payments(category);

-- Індекс для пошуку по періоду
CREATE INDEX IF NOT EXISTS idx_payments_period ON payments(period);

-- Індекс для пошуку по даті
CREATE INDEX IF NOT EXISTS idx_payments_date ON payments(payment_date);

-- Індекс для пошуку по користувачу
CREATE INDEX IF NOT EXISTS idx_payments_created_by ON payments(created_by);

-- Композитний індекс для типових запитів
CREATE INDEX IF NOT EXISTS idx_payments_apartment_period ON payments(apartment_id, period);

-- Композитний індекс для сортування
CREATE INDEX IF NOT EXISTS idx_payments_apartment_date ON payments(apartment_id, payment_date DESC);

-- Вставка тестових даних
-- Нарахування за січень 2025 для квартири 1
INSERT INTO payments (apartment_id, type, category, amount, description, payment_date, period, created_by, notes)
VALUES 
    (1, 'charge', 'maintenance', 500.00, 'Утримання будинку за січень 2025', '2025-01-01', '2025-01', 1, 'Базовий тариф'),
    (1, 'charge', 'utilities', 300.00, 'Комунальні послуги за січень 2025', '2025-01-01', '2025-01', 1, ''),
    (1, 'incoming', 'maintenance', 500.00, 'Оплата утримання будинку', '2025-01-15', '2025-01', 1, 'Платіжка №12345');

-- Нарахування для квартири 2 (з боргом)
INSERT INTO payments (apartment_id, type, category, amount, description, payment_date, period, created_by, notes)
VALUES 
    (2, 'charge', 'maintenance', 550.00, 'Утримання будинку за січень 2025', '2025-01-01', '2025-01', 1, ''),
    (2, 'charge', 'utilities', 320.00, 'Комунальні послуги за січень 2025', '2025-01-01', '2025-01', 1, ''),
    (2, 'incoming', 'maintenance', 300.00, 'Часткова оплата', '2025-01-20', '2025-01', 1, 'Платіжка №12346');

-- Нарахування для квартири 3 (повна оплата)
INSERT INTO payments (apartment_id, type, category, amount, description, payment_date, period, created_by, notes)
VALUES 
    (3, 'charge', 'maintenance', 700.00, 'Утримання будинку за січень 2025', '2025-01-01', '2025-01', 1, 'Велика квартира'),
    (3, 'charge', 'utilities', 400.00, 'Комунальні послуги за січень 2025', '2025-01-01', '2025-01', 1, ''),
    (3, 'incoming', 'maintenance', 700.00, 'Оплата утримання', '2025-01-10', '2025-01', 1, ''),
    (3, 'incoming', 'utilities', 400.00, 'Оплата комунальних', '2025-01-10', '2025-01', 1, '');

-- Примітка: Тестові дані показують різні сценарії:
-- - Квартира 1: частково оплачена (борг 300грн)
-- - Квартира 2: велика заборгованість (борг 570грн)
-- - Квартира 3: повністю оплачена (борг 0грн)