-- Міграція 001: Початкова схема БД для системи обліку ОСББ
-- Створює таблицю користувачів з підтримкою RBAC

-- Таблиця користувачів системи
CREATE TABLE IF NOT EXISTS users (
    -- Унікальний ідентифікатор користувача (auto-increment)
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- Унікальне ім'я для входу (case-sensitive)
    username TEXT NOT NULL UNIQUE,
    
    -- BCrypt хеш пароля (60 символів для BCrypt)
    hashed_password TEXT NOT NULL,
    
    -- Роль користувача (admin, head, accountant)
    role TEXT NOT NULL CHECK(role IN ('admin', 'head', 'accountant')),
    
    -- Повне ім'я користувача для відображення
    full_name TEXT NOT NULL,
    
    -- Час створення запису
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Час останнього оновлення
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Прапорець активності (для м'якого видалення)
    is_active BOOLEAN NOT NULL DEFAULT 1
);

-- Індекс для швидкого пошуку за username (використовується при логіні)
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Індекс для швидкого пошуку за роллю
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Індекс для фільтрації активних користувачів
CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active);

-- Композитний індекс для швидкого пошуку активних користувачів за роллю
CREATE INDEX IF NOT EXISTS idx_users_role_active ON users(role, is_active);

-- Вставка дефолтного адміністратора
-- Username: admin
-- Password: admin123 (ВАЖЛИВО: змінити після першого входу!)
-- BCrypt хеш для "admin123" з cost=12
INSERT INTO users (username, hashed_password, role, full_name, is_active)
VALUES (
    'admin',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIpPVd.3Lu',
    'admin',
    'Системний Адміністратор',
    1
);

-- ВАЖЛИВО: Після першого запуску системи змініть пароль адміністратора!
-- Цей дефолтний обліковий запис створений тільки для початкового налаштування.