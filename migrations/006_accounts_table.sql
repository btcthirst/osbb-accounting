-- =============================================================================
-- Таблиця Особистих Рахунків
-- =============================================================================


CREATE TABLE IF NOT EXISTS personal_accounts (
    -- Унікальний ідентифікатор рахунку
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- Номер особистого рахунку (унікальний)
    account_number TEXT NOT NULL UNIQUE,
    
    -- ID квартири (зовнішній ключ, унікальний - один рахунок на квартиру)
    apartment_id INTEGER NOT NULL UNIQUE,
    
    -- ID власника рахунку (зовнішній ключ)
    owner_id INTEGER NOT NULL,
    
    -- Дата відкриття рахунку
    opened_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Дата закриття рахунку (NULL = активний)
    closed_at DATETIME,
    
    -- Поточний баланс рахунку
    -- Додатній = переплата, від'ємний = борг
    current_balance REAL NOT NULL DEFAULT 0,
    
    -- Прапорець активності
    is_active BOOLEAN NOT NULL DEFAULT 1,
    
    -- Примітки про рахунок
    notes TEXT,
    
    -- Час створення запису
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Час останнього оновлення
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Зовнішні ключі
    FOREIGN KEY (apartment_id) REFERENCES apartments(id) ON DELETE RESTRICT,
    FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE RESTRICT
);

-- Індекс для швидкого пошуку за номером рахунку
CREATE INDEX IF NOT EXISTS idx_personal_accounts_number ON personal_accounts(account_number);

-- Індекс для пошуку рахунків квартири
CREATE INDEX IF NOT EXISTS idx_personal_accounts_apartment ON personal_accounts(apartment_id);

-- Індекс для пошуку рахунків власника
CREATE INDEX IF NOT EXISTS idx_personal_accounts_owner ON personal_accounts(owner_id);

-- Індекс для фільтрації активних рахунків
CREATE INDEX IF NOT EXISTS idx_personal_accounts_active ON personal_accounts(is_active);

-- Індекс для пошуку рахунків з боргом
CREATE INDEX IF NOT EXISTS idx_personal_accounts_debt ON personal_accounts(current_balance) 
    WHERE current_balance < 0;