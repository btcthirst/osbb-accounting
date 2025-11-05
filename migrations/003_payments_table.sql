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