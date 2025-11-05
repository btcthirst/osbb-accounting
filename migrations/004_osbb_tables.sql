-- =============================================================================
-- Таблиця ОСББ (Організація)
-- =============================================================================

CREATE TABLE IF NOT EXISTS osbb (
    -- Унікальний ідентифікатор (завжди 1, тому що ОСББ одне)
    id INTEGER PRIMARY KEY CHECK(id = 1),
    
    -- Назва ОСББ
    name TEXT NOT NULL,
    
    -- Скорочена назва (опціонально)
    short_name TEXT,
    
    -- Повна адреса будинку
    address TEXT NOT NULL,
    
    -- Код ЄДРПОУ (8 цифр)
    edrpou TEXT NOT NULL UNIQUE CHECK(length(edrpou) = 8),
    
    -- Базовий тариф за утримання (грн/м²)
    base_rate REAL NOT NULL DEFAULT 0 CHECK(base_rate >= 0),
    
    -- ПІБ голови правління
    chairman_name TEXT NOT NULL,
    
    -- Контактний телефон голови
    chairman_phone TEXT NOT NULL,
    
    -- Email голови (опціонально)
    chairman_email TEXT,
    
    -- Назва банку
    bank_name TEXT NOT NULL,
    
    -- Розрахунковий рахунок (IBAN)
    bank_account TEXT NOT NULL,
    
    -- МФО банку
    mfo TEXT NOT NULL,
    
    -- Дата створення ОСББ
    founded_at DATETIME NOT NULL,
    
    -- Час створення запису
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Час останнього оновлення
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Індекс для перевірки унікальності ЄДРПОУ
CREATE INDEX IF NOT EXISTS idx_osbb_edrpou ON osbb(edrpou);