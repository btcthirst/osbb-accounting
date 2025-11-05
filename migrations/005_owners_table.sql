-- =============================================================================
-- Таблиця Власників
-- =============================================================================

CREATE TABLE IF NOT EXISTS owners (
    -- Унікальний ідентифікатор власника
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- Повне ПІБ власника
    full_name TEXT NOT NULL,
    
    -- Індивідуальний податковий номер (ІПН/РНОКПП) - 10 цифр
    tax_id TEXT NOT NULL UNIQUE CHECK(length(tax_id) = 10),
    
    -- Основний контактний телефон
    phone TEXT NOT NULL,
    
    -- Email адреса (опціонально)
    email TEXT,
    
    -- Додатковий телефон (опціонально)
    alternative_phone TEXT,
    
    -- Серія паспорта (опціонально)
    passport_series TEXT,
    
    -- Номер паспорта (опціонально)
    passport_number TEXT,
    
    -- Додаткові примітки
    notes TEXT,
    
    -- Прапорець активності (для м'якого видалення)
    is_active BOOLEAN NOT NULL DEFAULT 1,
    
    -- Час створення запису
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Час останнього оновлення
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Індекс для швидкого пошуку за ІПН
CREATE INDEX IF NOT EXISTS idx_owners_tax_id ON owners(tax_id);

-- Індекс для пошуку за ПІБ
CREATE INDEX IF NOT EXISTS idx_owners_full_name ON owners(full_name);

-- Індекс для фільтрації активних власників
CREATE INDEX IF NOT EXISTS idx_owners_active ON owners(is_active);

-- Full-text search для пошуку власників (опціонально)
-- CREATE VIRTUAL TABLE IF NOT EXISTS owners_fts USING fts5(full_name, tax_id, content=owners);
