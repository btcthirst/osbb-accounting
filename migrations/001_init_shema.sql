-- ============================================================================
-- МІГРАЦІЯ 001: Початкова схема ОСББ (ВИПРАВЛЕНА ВЕРСІЯ v2.0)
-- Сумісність: SQLite 3.35+
-- Дата: 2025-11-09
-- Автор: Senior Go Developer
-- 
-- Зміни відносно попередньої версії:
--   ✅ Видалено циклічну залежність users ↔ owners
--   ✅ INTEGER timestamps замість TEXT (UNIX epoch)
--   ✅ Proper soft delete pattern (deleted_at)
--   ✅ Видалено created_by/updated_by з domain entities
--   ✅ Аудит перенесено повністю в audit_log
--   ✅ Покращена нормалізація
--   ✅ Додано missing indexes
-- ============================================================================

-- ============================================================================
-- PRAGMA налаштування для оптимізації SQLite
-- ============================================================================
-- PRAGMA foreign_keys = ON;
-- RAGMA journal_mode = WAL;
-- PRAGMA synchronous = NORMAL;
-- PRAGMA cache_size = -64000;  -- 64MB cache
-- PRAGMA temp_store = MEMORY;

-- ============================================================================
-- 1. SECURITY LAYER (створюється першою для FK залежностей)
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Таблиця: Користувачі системи
-- Примітка: Не містить owner_id для уникнення циклічної залежності
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE CHECK(length(username) >= 3 AND length(username) <= 50),
    email TEXT NOT NULL UNIQUE CHECK(email LIKE '%_@__%.__%'),
    password_hash TEXT NOT NULL CHECK(length(password_hash) = 60), -- BCrypt завжди 60 символів
    first_name TEXT NOT NULL CHECK(length(first_name) >= 1),
    last_name TEXT NOT NULL CHECK(length(last_name) >= 1),
    middle_name TEXT,
    phone TEXT CHECK(phone IS NULL OR length(phone) >= 10),
    is_active INTEGER NOT NULL DEFAULT 1 CHECK(is_active IN (0, 1)),
    
    -- Soft Delete Pattern
    deleted_at INTEGER DEFAULT NULL,
    
    -- Timestamps (UNIX epoch seconds)
    last_login_at INTEGER,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- Індекси
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;

-- ----------------------------------------------------------------------------
-- Таблиця: Ролі
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE CHECK(length(name) >= 2),
    description TEXT,
    is_system INTEGER NOT NULL DEFAULT 0 CHECK(is_system IN (0, 1)),
    is_active INTEGER NOT NULL DEFAULT 1 CHECK(is_active IN (0, 1)),
    
    -- Soft Delete
    deleted_at INTEGER DEFAULT NULL,
    
    -- Timestamps
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_name ON roles(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_roles_is_system ON roles(is_system);

-- Початкові системні ролі
INSERT OR IGNORE INTO roles (id, name, description, is_system) VALUES
(1, 'admin', 'Адміністратор системи - повний доступ', 1),
(2, 'accountant', 'Бухгалтер - фінансові операції та звіти', 1),
(3, 'manager', 'Управитель - перегляд даних та базові операції', 1),
(4, 'viewer', 'Спостерігач - тільки читання', 1);

-- ----------------------------------------------------------------------------
-- Таблиця: Дозволи (Permissions)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL UNIQUE CHECK(length(code) >= 3),
    name TEXT NOT NULL,
    description TEXT,
    resource TEXT NOT NULL CHECK(length(resource) >= 2),
    action TEXT NOT NULL CHECK(action IN ('create', 'read', 'update', 'delete', 'approve', 'export', 'all')),
    is_active INTEGER NOT NULL DEFAULT 1 CHECK(is_active IN (0, 1)),
    
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_permissions_code ON permissions(code);
CREATE INDEX IF NOT EXISTS idx_permissions_resource_action ON permissions(resource, action);

-- Початкові дозволи (повний набір)
INSERT OR IGNORE INTO permissions (code, name, description, resource, action) VALUES
-- Власники
('owners.read', 'Перегляд власників', 'Доступ до списку та деталей власників', 'owners', 'read'),
('owners.create', 'Створення власників', 'Можливість додавати нових власників', 'owners', 'create'),
('owners.update', 'Редагування власників', 'Можливість змінювати дані власників', 'owners', 'update'),
('owners.delete', 'Видалення власників', 'Можливість деактивувати власників', 'owners', 'delete'),

-- Квартири
('apartments.read', 'Перегляд квартир', 'Доступ до списку квартир', 'apartments', 'read'),
('apartments.create', 'Створення квартир', 'Можливість додавати квартири', 'apartments', 'create'),
('apartments.update', 'Редагування квартир', 'Можливість змінювати дані квартир', 'apartments', 'update'),
('apartments.delete', 'Видалення квартир', 'Можливість деактивувати квартири', 'apartments', 'delete'),

-- Частки власності
('ownership.read', 'Перегляд часток', 'Доступ до інформації про власність', 'ownership', 'read'),
('ownership.create', 'Створення часток', 'Можливість реєструвати нову власність', 'ownership', 'create'),
('ownership.update', 'Редагування часток', 'Можливість змінювати частки власності', 'ownership', 'update'),
('ownership.delete', 'Видалення часток', 'Можливість видаляти записи власності', 'ownership', 'delete'),

-- Нарахування
('charges.read', 'Перегляд нарахувань', 'Доступ до нарахувань', 'charges', 'read'),
('charges.create', 'Створення нарахувань', 'Можливість додавати нарахування', 'charges', 'create'),
('charges.update', 'Редагування нарахувань', 'Можливість змінювати нарахування', 'charges', 'update'),
('charges.delete', 'Видалення нарахувань', 'Можливість видаляти нарахування', 'charges', 'delete'),

-- Платежі
('payments.read', 'Перегляд платежів', 'Доступ до платежів', 'payments', 'read'),
('payments.create', 'Створення платежів', 'Можливість реєструвати платежі', 'payments', 'create'),
('payments.update', 'Редагування платежів', 'Можливість змінювати платежі', 'payments', 'update'),
('payments.delete', 'Видалення платежів', 'Можливість видаляти платежі', 'payments', 'delete'),
('payments.approve', 'Затвердження платежів', 'Можливість затверджувати платежі', 'payments', 'approve'),

-- Витрати
('expenses.read', 'Перегляд витрат', 'Доступ до витрат ОСББ', 'expenses', 'read'),
('expenses.create', 'Створення витрат', 'Можливість додавати витрати', 'expenses', 'create'),
('expenses.update', 'Редагування витрат', 'Можливість змінювати витрати', 'expenses', 'update'),
('expenses.delete', 'Видалення витрат', 'Можливість видаляти витрати', 'expenses', 'delete'),
('expenses.approve', 'Затвердження витрат', 'Можливість затверджувати витрати', 'expenses', 'approve'),

-- Контрагенти
('contractors.read', 'Перегляд контрагентів', 'Доступ до списку контрагентів', 'contractors', 'read'),
('contractors.create', 'Створення контрагентів', 'Можливість додавати контрагентів', 'contractors', 'create'),
('contractors.update', 'Редагування контрагентів', 'Можливість змінювати контрагентів', 'contractors', 'update'),
('contractors.delete', 'Видалення контрагентів', 'Можливість видаляти контрагентів', 'contractors', 'delete'),

-- Звіти
('reports.export', 'Експорт звітів', 'Можливість експортувати звіти', 'reports', 'export'),
('reports.read', 'Перегляд звітів', 'Доступ до перегляду звітів', 'reports', 'read'),

-- Користувачі (адмін функції)
('users.read', 'Перегляд користувачів', 'Доступ до списку користувачів', 'users', 'read'),
('users.create', 'Створення користувачів', 'Можливість додавати користувачів', 'users', 'create'),
('users.update', 'Редагування користувачів', 'Можливість змінювати користувачів', 'users', 'update'),
('users.delete', 'Видалення користувачів', 'Можливість видаляти користувачів', 'users', 'delete'),

-- Система
('system.all', 'Системний доступ', 'Повний доступ до всіх функцій', 'system', 'all');

-- ----------------------------------------------------------------------------
-- Таблиця: M:N зв'язок ролей та дозволів
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS role_permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    role_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    granted_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE,
    
    UNIQUE(role_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission ON role_permissions(permission_id);

-- Призначення дозволів системним ролям
-- Admin - все
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT 1, id FROM permissions WHERE code = 'system.all';

-- Accountant - фінансові операції
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions WHERE code IN (
    'owners.read', 'apartments.read', 'ownership.read',
    'charges.read', 'charges.create', 'charges.update', 'charges.delete',
    'payments.read', 'payments.create', 'payments.update', 'payments.approve',
    'expenses.read', 'expenses.create', 'expenses.update', 'expenses.approve',
    'contractors.read', 'contractors.create', 'contractors.update',
    'reports.export', 'reports.read'
);

-- Manager - перегляд та базові операції
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT 3, id FROM permissions WHERE code IN (
    'owners.read', 'apartments.read', 'ownership.read',
    'charges.read', 'payments.read',
    'expenses.read', 'contractors.read',
    'reports.read', 'reports.export'
);

-- Viewer - тільки читання
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT 4, id FROM permissions WHERE code IN (
    'owners.read', 'apartments.read', 'ownership.read',
    'charges.read', 'payments.read', 'reports.read'
);

-- ----------------------------------------------------------------------------
-- Таблиця: M:N зв'язок користувачів та ролей
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    granted_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    granted_by INTEGER,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (granted_by) REFERENCES users(id) ON DELETE SET NULL,
    
    UNIQUE(user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_user_roles_user ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_granted_by ON user_roles(granted_by);

-- ============================================================================
-- 2. DOMAIN LAYER - CORE ENTITIES (чисті від інфраструктури)
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Таблиця: Організація ОСББ (singleton record)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS osbb (
    id INTEGER PRIMARY KEY CHECK(id = 1), -- Тільки один запис
    name TEXT NOT NULL CHECK(length(name) >= 3),
    edrpou TEXT UNIQUE NOT NULL CHECK(length(edrpou) = 8 AND edrpou GLOB '[0-9]*'),
    legal_address TEXT NOT NULL,
    actual_address TEXT,
    phone TEXT CHECK(phone IS NULL OR length(phone) >= 10),
    email TEXT CHECK(email IS NULL OR email LIKE '%_@__%.__%'),
    website TEXT,
    chairman_name TEXT NOT NULL,
    
    -- Timestamps
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_osbb_edrpou ON osbb(edrpou);

-- ----------------------------------------------------------------------------
-- Таблиця: Власники квартир
-- Примітка: НЕ містить created_by/updated_by - аудит в окремій таблиці
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS owners (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT NOT NULL CHECK(length(first_name) >= 1),
    last_name TEXT NOT NULL CHECK(length(last_name) >= 1),
    middle_name TEXT,
    phone TEXT CHECK(phone IS NULL OR length(phone) >= 10),
    email TEXT CHECK(email IS NULL OR email LIKE '%_@__%.__%'),
    tax_number TEXT UNIQUE CHECK(tax_number IS NULL OR (length(tax_number) = 10 AND tax_number GLOB '[0-9]*')),
    passport_series TEXT,
    passport_number TEXT,
    registered_address TEXT,
    actual_address TEXT,
    notes TEXT,
    is_active INTEGER NOT NULL DEFAULT 1 CHECK(is_active IN (0, 1)),
    
    -- Soft Delete Pattern
    deleted_at INTEGER DEFAULT NULL,
    
    -- Timestamps
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- Індекси для швидкого пошуку
CREATE INDEX IF NOT EXISTS idx_owners_last_name ON owners(last_name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_owners_tax_number ON owners(tax_number) WHERE tax_number IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_owners_is_active ON owners(is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_owners_deleted_at ON owners(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_owners_full_name ON owners(last_name, first_name) WHERE deleted_at IS NULL;

-- ----------------------------------------------------------------------------
-- Таблиця: Квартири
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS apartments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    apartment_number TEXT NOT NULL UNIQUE CHECK(length(apartment_number) >= 1),
    floor INTEGER NOT NULL CHECK(floor >= 0),
    entrance INTEGER CHECK(entrance IS NULL OR entrance > 0),
    area_total REAL NOT NULL CHECK(area_total > 0),
    area_living REAL CHECK(area_living IS NULL OR (area_living > 0 AND area_living <= area_total)),
    rooms_count INTEGER CHECK(rooms_count IS NULL OR rooms_count > 0),
    cadastral_number TEXT UNIQUE,
    notes TEXT,
    is_active INTEGER NOT NULL DEFAULT 1 CHECK(is_active IN (0, 1)),
    
    -- Soft Delete
    deleted_at INTEGER DEFAULT NULL,
    
    -- Timestamps
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- Індекси
CREATE UNIQUE INDEX IF NOT EXISTS idx_apartments_number ON apartments(apartment_number) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_apartments_floor ON apartments(floor) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_apartments_is_active ON apartments(is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_apartments_deleted_at ON apartments(deleted_at) WHERE deleted_at IS NOT NULL;

-- ----------------------------------------------------------------------------
-- Таблиця: Частки власності (M:N зв'язок Owner ↔ Apartment)
-- Бізнес-правило: Один власник може мати кілька часток у різних квартирах
--                 Одна квартира може мати кілька власників
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ownership_shares (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id INTEGER NOT NULL,
    apartment_id INTEGER NOT NULL,
    
    -- Частка у форматі дріб (наприклад, 1/2, 1/4)
    share_numerator INTEGER NOT NULL CHECK(share_numerator > 0),
    share_denominator INTEGER NOT NULL CHECK(share_denominator > 0 AND share_denominator >= share_numerator),
    
    -- Тип власності
    ownership_type TEXT NOT NULL CHECK(ownership_type IN ('full', 'shared', 'rent')),
    
    -- Період дійсності
    start_date INTEGER NOT NULL,
    end_date INTEGER CHECK(end_date IS NULL OR end_date > start_date),
    
    -- Документальне підтвердження
    document_type TEXT CHECK(document_type IN ('contract', 'deed', 'certificate', 'court_decision', 'other')),
    document_number TEXT,
    document_date INTEGER,
    
    notes TEXT,
    is_active INTEGER NOT NULL DEFAULT 1 CHECK(is_active IN (0, 1)),
    
    -- Soft Delete
    deleted_at INTEGER DEFAULT NULL,
    
    -- Timestamps
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    
    -- Foreign Keys
    FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE RESTRICT,
    FOREIGN KEY (apartment_id) REFERENCES apartments(id) ON DELETE RESTRICT
);

-- Індекси для швидких зворотніх пошуків
CREATE INDEX IF NOT EXISTS idx_ownership_owner ON ownership_shares(owner_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ownership_apartment ON ownership_shares(apartment_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ownership_active ON ownership_shares(is_active, start_date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ownership_dates ON ownership_shares(start_date, end_date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ownership_deleted_at ON ownership_shares(deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================================
-- 3. FINANCIAL ENTITIES
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Таблиця: Контрагенти (постачальники послуг, підрядники)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS contractors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL CHECK(length(name) >= 2),
    edrpou TEXT UNIQUE CHECK(edrpou IS NULL OR (length(edrpou) BETWEEN 8 AND 10 AND edrpou GLOB '[0-9]*')),
    contractor_type TEXT NOT NULL CHECK(contractor_type IN ('utility', 'service', 'supplier', 'other')),
    contact_person TEXT,
    phone TEXT CHECK(phone IS NULL OR length(phone) >= 10),
    email TEXT CHECK(email IS NULL OR email LIKE '%_@__%.__%'),
    address TEXT,
    
    -- Банківські реквізити
    bank_account TEXT CHECK(bank_account IS NULL OR length(bank_account) = 29),
    bank_name TEXT,
    bank_mfo TEXT CHECK(bank_mfo IS NULL OR (length(bank_mfo) = 6 AND bank_mfo GLOB '[0-9]*')),
    
    -- Договірна інформація
    contract_number TEXT,
    contract_date INTEGER,
    
    notes TEXT,
    is_active INTEGER NOT NULL DEFAULT 1 CHECK(is_active IN (0, 1)),
    
    -- Soft Delete
    deleted_at INTEGER DEFAULT NULL,
    
    -- Timestamps
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_contractors_type ON contractors(contractor_type, is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractors_edrpou ON contractors(edrpou) WHERE edrpou IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractors_name ON contractors(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractors_deleted_at ON contractors(deleted_at) WHERE deleted_at IS NOT NULL;

-- ----------------------------------------------------------------------------
-- Таблиця: Категорії витрат (ієрархічна структура)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS expense_categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE CHECK(length(name) >= 2),
    category_type TEXT NOT NULL CHECK(category_type IN ('utility', 'repair', 'salary', 'service', 'other')),
    parent_id INTEGER,
    code TEXT UNIQUE CHECK(code IS NULL OR length(code) <= 20),
    description TEXT,
    is_active INTEGER NOT NULL DEFAULT 1 CHECK(is_active IN (0, 1)),
    
    -- Soft Delete
    deleted_at INTEGER DEFAULT NULL,
    
    -- Timestamps
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    
    FOREIGN KEY (parent_id) REFERENCES expense_categories(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_expense_categories_type ON expense_categories(category_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_expense_categories_parent ON expense_categories(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_expense_categories_code ON expense_categories(code) WHERE code IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_expense_categories_deleted_at ON expense_categories(deleted_at) WHERE deleted_at IS NOT NULL;

-- ----------------------------------------------------------------------------
-- Таблиця: Витрати ОСББ
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS expenses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL,
    contractor_id INTEGER,
    expense_date INTEGER NOT NULL,
    amount REAL NOT NULL CHECK(amount > 0),
    description TEXT NOT NULL CHECK(length(description) >= 3),
    
    -- Документальне підтвердження
    document_type TEXT CHECK(document_type IN ('invoice', 'act', 'receipt', 'order', 'other')),
    document_number TEXT,
    document_date INTEGER,
    
    -- Статус оплати
    payment_status TEXT NOT NULL DEFAULT 'pending' CHECK(payment_status IN ('pending', 'paid', 'partially_paid', 'cancelled')),
    paid_amount REAL DEFAULT 0 CHECK(paid_amount >= 0 AND paid_amount <= amount),
    payment_date INTEGER,
    
    notes TEXT,
    
    -- Затвердження (окремо від аудиту для бізнес-логіки)
    approved_by INTEGER,
    approved_at INTEGER,
    
    -- Soft Delete
    deleted_at INTEGER DEFAULT NULL,
    
    -- Timestamps
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    
    FOREIGN KEY (category_id) REFERENCES expense_categories(id) ON DELETE RESTRICT,
    FOREIGN KEY (contractor_id) REFERENCES contractors(id) ON DELETE RESTRICT,
    FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_expenses_date ON expenses(expense_date DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses(category_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_expenses_contractor ON expenses(contractor_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_expenses_status ON expenses(payment_status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_expenses_approved_by ON expenses(approved_by);
CREATE INDEX IF NOT EXISTS idx_expenses_deleted_at ON expenses(deleted_at) WHERE deleted_at IS NOT NULL;

-- ----------------------------------------------------------------------------
-- Таблиця: Нарахування (по частках власності)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS charges (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ownership_share_id INTEGER NOT NULL,
    charge_type TEXT NOT NULL CHECK(charge_type IN ('maintenance', 'utility', 'repair', 'penalty', 'other')),
    charge_date INTEGER NOT NULL,
    
    -- Період нарахування
    period_month INTEGER NOT NULL CHECK(period_month BETWEEN 1 AND 12),
    period_year INTEGER NOT NULL CHECK(period_year >= 2020 AND period_year <= 2100),
    
    -- Сума та розрахунок
    amount REAL NOT NULL CHECK(amount >= 0),
    tariff REAL CHECK(tariff IS NULL OR tariff > 0),
    quantity REAL CHECK(quantity IS NULL OR quantity > 0),
    
    description TEXT,
    notes TEXT,
    
    -- Soft Delete
    deleted_at INTEGER DEFAULT NULL,
    
    -- Timestamps
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    
    FOREIGN KEY (ownership_share_id) REFERENCES ownership_shares(id) ON DELETE RESTRICT,
    
    -- Унікальність: одне нарахування певного типу за період на частку
    UNIQUE(ownership_share_id, charge_type, period_year, period_month)
);

CREATE INDEX IF NOT EXISTS idx_charges_ownership ON charges(ownership_share_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_charges_period ON charges(period_year DESC, period_month DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_charges_type ON charges(charge_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_charges_date ON charges(charge_date DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_charges_deleted_at ON charges(deleted_at) WHERE deleted_at IS NOT NULL;

-- ----------------------------------------------------------------------------
-- Таблиця: Платежі власників
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS payments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ownership_share_id INTEGER NOT NULL,
    payment_date INTEGER NOT NULL,
    amount REAL NOT NULL CHECK(amount > 0),
    payment_method TEXT NOT NULL CHECK(payment_method IN ('cash', 'card', 'bank_transfer', 'other')),
    payment_purpose TEXT,
    
    -- Період, за який здійснюється платіж (опціонально)
    period_month INTEGER CHECK(period_month IS NULL OR period_month BETWEEN 1 AND 12),
    period_year INTEGER CHECK(period_year IS NULL OR (period_year >= 2020 AND period_year <= 2100)),
    
    receipt_number TEXT,
    notes TEXT,
    
    -- Затвердження платежу (для бізнес-логіки)
    approved_by INTEGER,
    approved_at INTEGER,
    
    -- Soft Delete
    deleted_at INTEGER DEFAULT NULL,
    
    -- Timestamps
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    
    FOREIGN KEY (ownership_share_id) REFERENCES ownership_shares(id) ON DELETE RESTRICT,
    FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_payments_ownership ON payments(ownership_share_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_payments_date ON payments(payment_date DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_payments_period ON payments(period_year DESC, period_month DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_payments_approved_by ON payments(approved_by);
CREATE INDEX IF NOT EXISTS idx_payments_deleted_at ON payments(deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================================
-- 4. USER-OWNER LINKING (розв'язання циклічної залежності)
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Таблиця: Зв'язок користувачів та власників (M:1)
-- Примітка: Замість owner_id в таблиці users
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_owner_links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE, -- Один користувач = один власник
    owner_id INTEGER NOT NULL,
    linked_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    linked_by INTEGER, -- Хто створив зв'язок
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE CASCADE,
    FOREIGN KEY (linked_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_owner_links_user ON user_owner_links(user_id);
CREATE INDEX IF NOT EXISTS idx_user_owner_links_owner ON user_owner_links(owner_id);
CREATE INDEX IF NOT EXISTS idx_user_owner_links_linked_by ON user_owner_links(linked_by);

-- ============================================================================
-- 5. AUDIT TRAIL (повний аудит всіх операцій)
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Таблиця: Журнал аудиту
-- Зберігає ВСІ зміни в системі включно з created_by/updated_by
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER, -- Хто виконав операцію
    table_name TEXT NOT NULL CHECK(length(table_name) >= 2),
    record_id INTEGER NOT NULL,
    action TEXT NOT NULL CHECK(action IN ('INSERT', 'UPDATE', 'DELETE', 'SOFT_DELETE', 'RESTORE')),
    
    -- JSON для зберігання старих/нових значень
    old_values TEXT, -- JSON object
    new_values TEXT, -- JSON object
    
    -- Додаткова інформація
    ip_address TEXT,
    user_agent TEXT,
    notes TEXT,
    
    -- Timestamp
    changed_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
    
    -- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
    -- Намисно закоментовано щоб audit_log був незалежним
);

CREATE INDEX IF NOT EXISTS idx_audit_log_user ON audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_table_record ON audit_log(table_name, record_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_date ON audit_log(changed_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_action ON audit_log(action);

-- ============================================================================
-- 6. TRIGGERS - AUTOMATIC UPDATES
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Тригери для автоматичного оновлення updated_at
-- ----------------------------------------------------------------------------

-- OSBB
CREATE TRIGGER IF NOT EXISTS trg_osbb_updated_at 
AFTER UPDATE ON osbb
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE osbb SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
END;

-- Users
CREATE TRIGGER IF NOT EXISTS trg_users_updated_at 
AFTER UPDATE ON users
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at AND NEW.deleted_at IS NULL
BEGIN
    UPDATE users SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
END;

-- Owners
CREATE TRIGGER IF NOT EXISTS trg_owners_updated_at 
AFTER UPDATE ON owners
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at AND NEW.deleted_at IS NULL
BEGIN
    UPDATE owners SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
END;

-- Apartments
CREATE TRIGGER IF NOT EXISTS trg_apartments_updated_at 
AFTER UPDATE ON apartments
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at AND NEW.deleted_at IS NULL
BEGIN
    UPDATE apartments SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
END;

-- Ownership Shares
CREATE TRIGGER IF NOT EXISTS trg_ownership_shares_updated_at 
AFTER UPDATE ON ownership_shares
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at AND NEW.deleted_at IS NULL
BEGIN
    UPDATE ownership_shares SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
END;

-- Contractors
CREATE TRIGGER IF NOT EXISTS trg_contractors_updated_at 
AFTER UPDATE ON contractors
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at AND NEW.deleted_at IS NULL
BEGIN
    UPDATE contractors SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
END;

-- Expense Categories
CREATE TRIGGER IF NOT EXISTS trg_expense_categories_updated_at 
AFTER UPDATE ON expense_categories
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at AND NEW.deleted_at IS NULL
BEGIN
    UPDATE expense_categories SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
END;

-- Expenses
CREATE TRIGGER IF NOT EXISTS trg_expenses_updated_at 
AFTER UPDATE ON expenses
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at AND NEW.deleted_at IS NULL
BEGIN
    UPDATE expenses SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
END;

-- Charges
CREATE TRIGGER IF NOT EXISTS trg_charges_updated_at 
AFTER UPDATE ON charges
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at AND NEW.deleted_at IS NULL
BEGIN
    UPDATE charges SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
END;

-- Payments
CREATE TRIGGER IF NOT EXISTS trg_payments_updated_at 
AFTER UPDATE ON payments
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at AND NEW.deleted_at IS NULL
BEGIN
    UPDATE payments SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
END;

-- Roles
CREATE TRIGGER IF NOT EXISTS trg_roles_updated_at 
AFTER UPDATE ON roles
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at AND NEW.deleted_at IS NULL
BEGIN
    UPDATE roles SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
END;

-- ============================================================================
-- 7. BUSINESS RULES TRIGGERS
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Тригер: Заборона дублювання активних часток власності
-- ----------------------------------------------------------------------------
CREATE TRIGGER IF NOT EXISTS trg_prevent_duplicate_ownership
BEFORE INSERT ON ownership_shares
FOR EACH ROW
WHEN NEW.is_active = 1 
  AND NEW.deleted_at IS NULL
  AND EXISTS (
    SELECT 1 FROM ownership_shares 
    WHERE owner_id = NEW.owner_id 
      AND apartment_id = NEW.apartment_id 
      AND is_active = 1
      AND deleted_at IS NULL
      AND id != COALESCE(NEW.id, 0)
      AND (end_date IS NULL OR end_date > NEW.start_date)
  )
BEGIN
    SELECT RAISE(ABORT, 'Власник вже має активну частку в цій квартирі у вказаний період');
END;

-- ----------------------------------------------------------------------------
-- Тригер: Валідація суми часток у квартирі (не більше 100%)
-- ----------------------------------------------------------------------------
CREATE TRIGGER IF NOT EXISTS trg_validate_ownership_sum
BEFORE INSERT ON ownership_shares
FOR EACH ROW
WHEN NEW.is_active = 1 AND NEW.deleted_at IS NULL
BEGIN
    SELECT CASE
        WHEN (
            SELECT CAST(SUM(CAST(share_numerator AS REAL) / share_denominator) AS REAL)
            FROM ownership_shares
            WHERE apartment_id = NEW.apartment_id
              AND is_active = 1
              AND deleted_at IS NULL
              AND id != COALESCE(NEW.id, 0)
        ) + (CAST(NEW.share_numerator AS REAL) / NEW.share_denominator) > 1.0001 -- Невелика похибка для float
        THEN RAISE(ABORT, 'Сума часток власності у квартирі перевищує 100%')
    END;
END;

-- ----------------------------------------------------------------------------
-- Тригер: Автоматичне завершення старих часток при створенні нових
-- (опціонально - коментуємо для ручного контролю)
-- ----------------------------------------------------------------------------
/*
CREATE TRIGGER IF NOT EXISTS trg_deactivate_old_ownership
AFTER INSERT ON ownership_shares
FOR EACH ROW
WHEN NEW.is_active = 1 AND NEW.deleted_at IS NULL
BEGIN
    UPDATE ownership_shares 
    SET is_active = 0, 
        end_date = NEW.start_date,
        updated_at = strftime('%s', 'now')
    WHERE owner_id = NEW.owner_id 
      AND apartment_id = NEW.apartment_id 
      AND is_active = 1
      AND deleted_at IS NULL
      AND id != NEW.id
      AND start_date < NEW.start_date;
END;
*/

-- ----------------------------------------------------------------------------
-- Тригер: Валідація платежу (не може бути більшим за залишок боргу)
-- (опціонально - для суворого контролю)
-- ----------------------------------------------------------------------------
/*
CREATE TRIGGER IF NOT EXISTS trg_validate_payment_amount
BEFORE INSERT ON payments
FOR EACH ROW
BEGIN
    SELECT CASE
        WHEN NEW.amount > (
            SELECT COALESCE(SUM(c.amount), 0) - COALESCE(SUM(p.amount), 0)
            FROM ownership_shares os
            LEFT JOIN charges c ON c.ownership_share_id = os.id AND c.deleted_at IS NULL
            LEFT JOIN payments p ON p.ownership_share_id = os.id AND p.deleted_at IS NULL AND p.id != NEW.id
            WHERE os.id = NEW.ownership_share_id
        )
        THEN RAISE(ABORT, 'Сума платежу перевищує залишок боргу')
    END;
END;
*/

-- ============================================================================
-- 8. VIEWS - COMPLEX QUERIES
-- ============================================================================

-- ----------------------------------------------------------------------------
-- View: Повна інформація про власність
-- ----------------------------------------------------------------------------
CREATE VIEW IF NOT EXISTS v_ownership_full AS
SELECT 
    os.id AS ownership_id,
    os.owner_id,
    o.last_name || ' ' || o.first_name || COALESCE(' ' || o.middle_name, '') AS owner_full_name,
    o.phone AS owner_phone,
    o.email AS owner_email,
    o.tax_number AS owner_tax_number,
    os.apartment_id,
    a.apartment_number,
    a.floor,
    a.entrance,
    a.area_total,
    a.area_living,
    a.rooms_count,
    os.share_numerator,
    os.share_denominator,
    os.share_numerator || '/' || os.share_denominator AS share_fraction,
    ROUND(CAST(os.share_numerator AS REAL) / os.share_denominator * 100, 2) AS share_percent,
    ROUND(a.area_total * CAST(os.share_numerator AS REAL) / os.share_denominator, 2) AS area_owned,
    os.ownership_type,
    os.start_date,
    os.end_date,
    os.is_active,
    os.document_type,
    os.document_number,
    os.created_at,
    os.updated_at
FROM ownership_shares os
JOIN owners o ON os.owner_id = o.id
JOIN apartments a ON os.apartment_id = a.id
WHERE os.deleted_at IS NULL
  AND o.deleted_at IS NULL
  AND a.deleted_at IS NULL;

-- ----------------------------------------------------------------------------
-- View: Баланс власників (нарахування - платежі)
-- ----------------------------------------------------------------------------
CREATE VIEW IF NOT EXISTS v_owner_balances AS
SELECT 
    os.id AS ownership_share_id,
    os.owner_id,
    o.last_name || ' ' || o.first_name AS owner_name,
    os.apartment_id,
    a.apartment_number,
    COALESCE(SUM(c.amount), 0) AS total_charged,
    COALESCE(SUM(p.amount), 0) AS total_paid,
    COALESCE(SUM(c.amount), 0) - COALESCE(SUM(p.amount), 0) AS balance,
    CASE 
        WHEN COALESCE(SUM(c.amount), 0) - COALESCE(SUM(p.amount), 0) > 0 THEN 'debtor'
        WHEN COALESCE(SUM(c.amount), 0) - COALESCE(SUM(p.amount), 0) < 0 THEN 'overpaid'
        ELSE 'balanced'
    END AS balance_status,
    MAX(c.charge_date) AS last_charge_date,
    MAX(p.payment_date) AS last_payment_date
FROM ownership_shares os
LEFT JOIN owners o ON os.owner_id = o.id
LEFT JOIN apartments a ON os.apartment_id = a.id
LEFT JOIN charges c ON c.ownership_share_id = os.id AND c.deleted_at IS NULL
LEFT JOIN payments p ON p.ownership_share_id = os.id AND p.deleted_at IS NULL
WHERE os.is_active = 1
  AND os.deleted_at IS NULL
  AND o.deleted_at IS NULL
  AND a.deleted_at IS NULL
GROUP BY os.id, os.owner_id, os.apartment_id;

-- ----------------------------------------------------------------------------
-- View: Баланс по періодах (місячний розріз)
-- ----------------------------------------------------------------------------
CREATE VIEW IF NOT EXISTS v_period_balances AS
SELECT 
    os.id AS ownership_share_id,
    os.owner_id,
    os.apartment_id,
    c.period_year,
    c.period_month,
    COALESCE(SUM(c.amount), 0) AS charged,
    COALESCE(SUM(p.amount), 0) AS paid,
    COALESCE(SUM(c.amount), 0) - COALESCE(SUM(p.amount), 0) AS period_balance
FROM ownership_shares os
LEFT JOIN charges c ON c.ownership_share_id = os.id AND c.deleted_at IS NULL
LEFT JOIN payments p ON p.ownership_share_id = os.id 
    AND p.period_year = c.period_year 
    AND p.period_month = c.period_month
    AND p.deleted_at IS NULL
WHERE os.deleted_at IS NULL
GROUP BY os.id, c.period_year, c.period_month
ORDER BY c.period_year DESC, c.period_month DESC;

-- ----------------------------------------------------------------------------
-- View: Користувачі з ролями та дозволами
-- ----------------------------------------------------------------------------
CREATE VIEW IF NOT EXISTS v_user_permissions AS
SELECT DISTINCT
    u.id AS user_id,
    u.username,
    u.email,
    u.first_name,
    u.last_name,
    u.is_active AS user_is_active,
    r.id AS role_id,
    r.name AS role_name,
    r.description AS role_description,
    p.id AS permission_id,
    p.code AS permission_code,
    p.name AS permission_name,
    p.resource,
    p.action
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
WHERE u.is_active = 1 
  AND u.deleted_at IS NULL
  AND r.is_active = 1 
  AND r.deleted_at IS NULL
  AND p.is_active = 1;

-- ----------------------------------------------------------------------------
-- View: Користувачі зв'язані з власниками
-- ----------------------------------------------------------------------------
CREATE VIEW IF NOT EXISTS v_user_owners AS
SELECT 
    u.id AS user_id,
    u.username,
    u.email,
    u.first_name AS user_first_name,
    u.last_name AS user_last_name,
    u.is_active AS user_is_active,
    uol.owner_id,
    o.first_name AS owner_first_name,
    o.last_name AS owner_last_name,
    o.phone AS owner_phone,
    o.email AS owner_email,
    o.is_active AS owner_is_active,
    uol.linked_at
FROM users u
LEFT JOIN user_owner_links uol ON u.id = uol.user_id
LEFT JOIN owners o ON uol.owner_id = o.id
WHERE u.deleted_at IS NULL
  AND (o.id IS NULL OR o.deleted_at IS NULL);

-- ----------------------------------------------------------------------------
-- View: Статистика по квартирах
-- ----------------------------------------------------------------------------
CREATE VIEW IF NOT EXISTS v_apartment_stats AS
SELECT 
    a.id AS apartment_id,
    a.apartment_number,
    a.floor,
    a.area_total,
    COUNT(DISTINCT os.owner_id) AS owners_count,
    COALESCE(SUM(CAST(os.share_numerator AS REAL) / os.share_denominator), 0) AS total_ownership_ratio,
    ROUND(
        COALESCE(SUM(c.amount), 0) - COALESCE(SUM(p.amount), 0), 
        2
    ) AS total_debt,
    MAX(c.charge_date) AS last_charge_date,
    MAX(p.payment_date) AS last_payment_date
FROM apartments a
LEFT JOIN ownership_shares os ON a.id = os.apartment_id 
    AND os.is_active = 1 
    AND os.deleted_at IS NULL
LEFT JOIN charges c ON os.id = c.ownership_share_id 
    AND c.deleted_at IS NULL
LEFT JOIN payments p ON os.id = p.ownership_share_id 
    AND p.deleted_at IS NULL
WHERE a.deleted_at IS NULL
GROUP BY a.id;

-- ----------------------------------------------------------------------------
-- View: Витрати з категоріями та контрагентами
-- ----------------------------------------------------------------------------
CREATE VIEW IF NOT EXISTS v_expenses_full AS
SELECT 
    e.id AS expense_id,
    e.expense_date,
    e.amount,
    e.description,
    e.payment_status,
    e.paid_amount,
    e.payment_date,
    ec.name AS category_name,
    ec.category_type,
    ec.code AS category_code,
    c.name AS contractor_name,
    c.edrpou AS contractor_edrpou,
    c.contractor_type,
    e.document_type,
    e.document_number,
    e.document_date,
    e.approved_by,
    e.approved_at,
    u.username AS approved_by_username,
    e.created_at,
    e.updated_at
FROM expenses e
JOIN expense_categories ec ON e.category_id = ec.id
LEFT JOIN contractors c ON e.contractor_id = c.id
LEFT JOIN users u ON e.approved_by = u.id
WHERE e.deleted_at IS NULL
  AND ec.deleted_at IS NULL
  AND (c.id IS NULL OR c.deleted_at IS NULL);

-- ============================================================================
-- 9. INITIAL DATA / SEED
-- ============================================================================

-- ----------------------------------------------------------------------------
-- OSBB Organization (Singleton)
-- ----------------------------------------------------------------------------
INSERT OR IGNORE INTO osbb (
    id, 
    name, 
    edrpou, 
    legal_address, 
    chairman_name
) VALUES (
    1, 
    'ОСББ "Приклад"', 
    '12345678', 
    'м. Київ, вул. Центральна, 1',
    'Іванов Іван Іванович'
);

-- ----------------------------------------------------------------------------
-- Базові категорії витрат
-- ----------------------------------------------------------------------------
INSERT OR IGNORE INTO expense_categories (id, name, category_type, code, description) VALUES
-- Комунальні послуги
(1, 'Комунальні послуги', 'utility', 'UTIL', 'Загальна категорія комунальних послуг'),
(2, 'Електроенергія', 'utility', 'UTIL-ELEC', 'Електропостачання'),
(3, 'Водопостачання', 'utility', 'UTIL-WATER', 'Холодна та гаряча вода'),
(4, 'Газопостачання', 'utility', 'UTIL-GAS', 'Природний газ'),
(5, 'Опалення', 'utility', 'UTIL-HEAT', 'Центральне опалення'),

-- Ремонт та утримання
(10, 'Ремонт', 'repair', 'REPAIR', 'Загальна категорія ремонтних робіт'),
(11, 'Ремонт даху', 'repair', 'REPAIR-ROOF', 'Ремонт покрівлі'),
(12, 'Ремонт фасаду', 'repair', 'REPAIR-FACADE', 'Ремонт фасаду будинку'),
(13, 'Ремонт ліфтів', 'repair', 'REPAIR-LIFT', 'Обслуговування та ремонт ліфтів'),
(14, 'Благоустрій', 'repair', 'REPAIR-YARD', 'Благоустрій прибудинкової території'),

-- Зарплата
(20, 'Зарплата', 'salary', 'SALARY', 'Заробітна плата працівників'),
(21, 'Зарплата двірника', 'salary', 'SALARY-JANITOR', 'Двірник'),
(22, 'Зарплата консьєржа', 'salary', 'SALARY-CONCIERGE', 'Консьєрж'),

-- Послуги
(30, 'Послуги', 'service', 'SERVICE', 'Інші послуги'),
(31, 'Охорона', 'service', 'SERVICE-SECURITY', 'Послуги охорони'),
(32, 'Вивіз сміття', 'service', 'SERVICE-GARBAGE', 'Вивіз твердих побутових відходів'),
(33, 'Бухгалтерські послуги', 'service', 'SERVICE-ACCOUNTING', 'Послуги бухгалтера'),

-- Інше
(40, 'Інше', 'other', 'OTHER', 'Інші витрати');

-- Встановлюємо parent_id для ієрархії
UPDATE expense_categories SET parent_id = 1 WHERE id IN (2, 3, 4, 5);
UPDATE expense_categories SET parent_id = 10 WHERE id IN (11, 12, 13, 14);
UPDATE expense_categories SET parent_id = 20 WHERE id IN (21, 22);
UPDATE expense_categories SET parent_id = 30 WHERE id IN (31, 32, 33);

-- ============================================================================
-- 10. HELPER FUNCTIONS (для SQLite через application layer)
-- ============================================================================

-- Примітка: SQLite не підтримує користувацькі функції напряму в SQL.
-- Ці функції будуть реалізовані в Go коді:
--
-- 1. CalculateBalance(ownershipShareID) - розрахунок балансу
-- 2. CalculatePeriodBalance(ownershipShareID, year, month) - баланс за період
-- 3. GetUserPermissions(userID) - отримання всіх дозволів користувача
-- 4. HasPermission(userID, permissionCode) - перевірка конкретного дозволу
-- 5. GetActiveOwnershipShares(ownerID) - активні частки власника
-- 6. GetOwnersByApartment(apartmentID) - всі власники квартири
-- 7. SoftDelete(tableName, recordID, userID) - м'яке видалення з audit log

-- ============================================================================
-- 11. MIGRATION METADATA
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Таблиця: Версії міграцій (для tracking)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS schema_migrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version TEXT NOT NULL UNIQUE,
    description TEXT,
    applied_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    checksum TEXT -- MD5/SHA256 для верифікації
);

-- Реєструємо поточну міграцію
INSERT OR IGNORE INTO schema_migrations (version, description) VALUES
('001', 'Initial schema - Core entities, Security, Financial, Audit');

-- ============================================================================
-- 12. PERFORMANCE OPTIMIZATION HINTS
-- ============================================================================

-- Для великих БД рекомендується:
-- ANALYZE; -- Після заповнення даними для оптимізації query planner

-- Регулярне обслуговування:
-- VACUUM; -- Раз на місяць для дефрагментації
-- PRAGMA optimize; -- Після великих змін у даних

-- ============================================================================
-- КІНЕЦЬ МІГРАЦІЇ 001 v2.0
-- ============================================================================

-- Перевірка цілісності після міграції
-- PRAGMA foreign_key_check;
-- PRAGMA integrity_check;

-- Виведення статистики
SELECT 
    'Migration 001 v2.0 completed successfully' AS status,
    (SELECT COUNT(*) FROM sqlite_master WHERE type='table') AS tables_count,
    (SELECT COUNT(*) FROM sqlite_master WHERE type='index') AS indexes_count,
    (SELECT COUNT(*) FROM sqlite_master WHERE type='trigger') AS triggers_count,
    (SELECT COUNT(*) FROM sqlite_master WHERE type='view') AS views_count;