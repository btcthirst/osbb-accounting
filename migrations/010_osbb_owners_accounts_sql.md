-- Міграція 004: Таблиці ОСББ, Власників та Особистих Рахунків
-- Розширює систему для повноцінного обліку з фінансовими рахунками

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

-- =============================================================================
-- Оновлення таблиці квартир
-- =============================================================================

-- Додаємо зв'язок з власником в таблицю apartments
-- ВАЖЛИВО: Це буде виконано тільки якщо колонка не існує

-- Перевіряємо, чи існує колонка owner_id
-- Якщо ні - додаємо її
ALTER TABLE apartments ADD COLUMN owner_id INTEGER REFERENCES owners(id);

-- Індекс для пошуку квартир власника
CREATE INDEX IF NOT EXISTS idx_apartments_owner ON apartments(owner_id);

-- =============================================================================
-- Оновлення таблиці платежів
-- =============================================================================

-- Додаємо зв'язок з особистим рахунком в таблицю payments
ALTER TABLE payments ADD COLUMN personal_account_id INTEGER REFERENCES personal_accounts(id);

-- Індекс для пошуку платежів по особистому рахунку
CREATE INDEX IF NOT EXISTS idx_payments_personal_account ON payments(personal_account_id);

-- =============================================================================
-- Тригери для автоматичного оновлення updated_at
-- =============================================================================

-- Тригер для таблиці ОСББ
CREATE TRIGGER IF NOT EXISTS update_osbb_timestamp 
AFTER UPDATE ON osbb
BEGIN
    UPDATE osbb SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Тригер для таблиці власників
CREATE TRIGGER IF NOT EXISTS update_owners_timestamp 
AFTER UPDATE ON owners
BEGIN
    UPDATE owners SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Тригер для таблиці особистих рахунків
CREATE TRIGGER IF NOT EXISTS update_personal_accounts_timestamp 
AFTER UPDATE ON personal_accounts
BEGIN
    UPDATE personal_accounts SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- =============================================================================
-- View для зручного перегляду даних
-- =============================================================================

-- View: Квартири з власниками та рахунками
CREATE VIEW IF NOT EXISTS v_apartments_full AS
SELECT 
    a.*,
    o.full_name as owner_full_name,
    o.tax_id as owner_tax_id,
    o.phone as owner_phone,
    o.email as owner_email,
    pa.account_number,
    pa.current_balance,
    pa.opened_at as account_opened_at,
    CASE 
        WHEN pa.current_balance < 0 THEN 1 
        ELSE 0 
    END as has_debt,
    ABS(pa.current_balance) as debt_amount
FROM apartments a
LEFT JOIN owners o ON a.owner_id = o.id
LEFT JOIN personal_accounts pa ON a.id = pa.apartment_id
WHERE a.is_active = 1;

-- View: Власники з кількістю квартир та загальним балансом
CREATE VIEW IF NOT EXISTS v_owners_summary AS
SELECT 
    o.*,
    COUNT(DISTINCT pa.id) as accounts_count,
    COUNT(DISTINCT a.id) as apartments_count,
    COALESCE(SUM(pa.current_balance), 0) as total_balance,
    COALESCE(SUM(CASE WHEN pa.current_balance < 0 THEN pa.current_balance ELSE 0 END), 0) as total_debt,
    COALESCE(SUM(CASE WHEN pa.current_balance > 0 THEN pa.current_balance ELSE 0 END), 0) as total_overpayment
FROM owners o
LEFT JOIN personal_accounts pa ON o.id = pa.owner_id
LEFT JOIN apartments a ON o.id = a.owner_id
WHERE o.is_active = 1
GROUP BY o.id;

-- =============================================================================
-- Тестові дані (опціонально)
-- =============================================================================

-- Вставка даних ОСББ (приклад)
INSERT OR IGNORE INTO osbb (
    id, name, short_name, address, edrpou, base_rate,
    chairman_name, chairman_phone, chairman_email,
    bank_name, bank_account, mfo, founded_at
) VALUES (
    1,
    'Об''єднання Співвласників Багатоквартирного Будинку "Сонячний"',
    'ОСББ "Сонячний"',
    'м. Київ, вул. Центральна, 42',
    '12345678',
    15.50,
    'Іваненко Іван Іванович',
    '+380501234567',
    'chairman@osbb-sonyachnyy.ua',
    'ПриватБанк',
    'UA123456789012345678901234567',
    '305299',
    '2020-01-15'
);

-- Вставка тестових власників
INSERT OR IGNORE INTO owners (full_name, tax_id, phone, email)
VALUES 
    ('Іванов Іван Іванович', '1234567890', '+380501234567', 'ivanov@example.com'),
    ('Петренко Петро Петрович', '0987654321', '+380502345678', 'petrenko@example.com'),
    ('Сидоренко Сидір Сидорович', '1122334455', '+380503456789', NULL);

-- Оновлюємо існуючі квартири - прив'язуємо до власників
UPDATE apartments SET owner_id = 1 WHERE id = 1;
UPDATE apartments SET owner_id = 2 WHERE id = 2;
UPDATE apartments SET owner_id = 3 WHERE id = 3;

-- Створюємо особисті рахунки для існуючих квартир
INSERT OR IGNORE INTO personal_accounts (account_number, apartment_id, owner_id, current_balance)
VALUES 
    ('0001-0001', 1, 1, -300.00),  -- Борг 300 грн
    ('0001-0002', 2, 2, -570.00),  -- Борг 570 грн
    ('0001-0003', 3, 3, 0.00);     -- Без боргу

-- Примітка: Після цієї міграції система буде готова до роботи з 
-- особистими рахунками та фінансовим обліком на рівні ОСББ