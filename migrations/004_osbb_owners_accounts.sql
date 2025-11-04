-- Міграція 004: ОСББ, Власники, Особисті Рахунки
-- Створює таблиці для організації, власників та фінансового обліку

-- =====================================================
-- Таблиця ОСББ (єдина організація)
-- =====================================================
CREATE TABLE IF NOT EXISTS osbb (
    id INTEGER PRIMARY KEY CHECK(id = 1), -- Завжди 1
    name TEXT NOT NULL,
    short_name TEXT,
    address TEXT NOT NULL,
    edrpou TEXT,
    head_of_board TEXT,
    base_maintenance_rate REAL NOT NULL DEFAULT 15.0 CHECK(base_maintenance_rate >= 0),
    base_utilities_rate REAL NOT NULL DEFAULT 10.0 CHECK(base_utilities_rate >= 0),
    bank_name TEXT,
    bank_account TEXT,
    bank_mfo TEXT,
    phone TEXT,
    email TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Вставка дефолтних даних ОСББ
INSERT INTO osbb (id, name, short_name, address, head_of_board, base_maintenance_rate, base_utilities_rate)
VALUES (
    1,
    'ОСББ "Наш Дім"',
    'ОСББ "Наш Дім"',
    'м. Київ, вул. Хрещатик, буд. 1',
    'Іванов Іван Іванович',
    15.00,
    10.00
);

-- =====================================================
-- Таблиця власників (фізичні особи)
-- =====================================================
CREATE TABLE IF NOT EXISTS owners (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    full_name TEXT NOT NULL,
    tax_id TEXT NOT NULL UNIQUE, -- ІПН (10 цифр)
    phone TEXT NOT NULL,
    phone_additional TEXT,
    email TEXT,
    passport_series TEXT,
    passport_number TEXT,
    passport_issued_by TEXT,
    passport_issued_date DATE,
    registration_address TEXT,
    notes TEXT,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Індекси для owners
CREATE INDEX IF NOT EXISTS idx_owners_tax_id ON owners(tax_id);
CREATE INDEX IF NOT EXISTS idx_owners_phone ON owners(phone);
CREATE INDEX IF NOT EXISTS idx_owners_active ON owners(is_active);
CREATE INDEX IF NOT EXISTS idx_owners_name ON owners(full_name);

-- =====================================================
-- Таблиця особистих рахунків (зв'язок власник-квартира)
-- =====================================================
CREATE TABLE IF NOT EXISTS personal_accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_number TEXT NOT NULL UNIQUE, -- Унікальний номер ОР (10-12 цифр)
    apartment_id INTEGER NOT NULL,
    owner_id INTEGER NOT NULL,
    open_date DATE NOT NULL DEFAULT CURRENT_DATE,
    close_date DATE,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    notes TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (apartment_id) REFERENCES apartments(id) ON DELETE CASCADE,
    FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE RESTRICT
);

-- Індекси для personal_accounts
CREATE INDEX IF NOT EXISTS idx_pa_account_number ON personal_accounts(account_number);
CREATE INDEX IF NOT EXISTS idx_pa_apartment ON personal_accounts(apartment_id);
CREATE INDEX IF NOT EXISTS idx_pa_owner ON personal_accounts(owner_id);
CREATE INDEX IF NOT EXISTS idx_pa_active ON personal_accounts(is_active);
CREATE UNIQUE INDEX IF NOT EXISTS idx_pa_apartment_active ON personal_accounts(apartment_id, is_active) 
    WHERE is_active = 1;

-- =====================================================
-- Тестові дані
-- =====================================================

-- Власники (для існуючих квартир)
INSERT INTO owners (full_name, tax_id, phone, email, is_active)
VALUES 
    ('Іванов Іван Іванович', '1234567890', '+380501234567', 'ivanov@example.com', 1),
    ('Петренко Петро Петрович', '2345678901', '+380502345678', 'petrenko@example.com', 1),
    ('Сидоренко Сидір Сидорович', '3456789012', '+380503456789', 'sydorenko@example.com', 1),
    ('Коваленко Ольга Іванівна', '4567890123', '+380504567890', 'kovalenko@example.com', 1),
    ('Мельник Марія Петрівна', '5678901234', '+380505678901', 'melnyk@example.com', 1),
    ('Шевченко Тарас Григорович', '6789012345', '+380506789012', 'shevchenko@example.com', 1);

-- Особисті рахунки для існуючих квартир
-- Формат номера: 10001XXXXX (10001 + номер квартири)
INSERT INTO personal_accounts (account_number, apartment_id, owner_id, open_date, is_active)
VALUES 
    ('1000100001', 1, 1, '2024-01-01', 1), -- Кв. 1 -> Іванов
    ('1000100002', 2, 2, '2024-01-01', 1), -- Кв. 2 -> Петренко
    ('1000100003', 3, 3, '2024-01-01', 1), -- Кв. 3 -> Сидоренко
    ('1000100004', 4, 4, '2024-01-01', 1), -- Кв. 4 -> Коваленко
    ('1000100005', 5, 5, '2024-01-01', 1), -- Кв. 5 -> Мельник
    ('1000100006', 6, 6, '2024-01-01', 1); -- Кв. 6 -> Шевченко

-- =====================================================
-- Примітки
-- =====================================================
-- 1. ОСББ: Єдина організація (ID завжди = 1)
-- 2. Власники: Окрема таблиця фізосіб з повними даними
-- 3. Особисті Рахунки: Зв'язок власник-квартира для бухобліку
-- 4. Номер ОР: Унікальний код для ідентифікації у банку/бухгалтерії
-- 5. Одна квартира = один активний ОР
-- 6. Один власник може мати декілька квартир (декілька ОР)