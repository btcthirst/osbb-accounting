-- ============================================================================
-- МІГРАЦІЯ 002: Таблиця сесій
-- Дата: 2025-11-09
-- Опис: Додає таблицю для зберігання сесій користувачів
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Таблиця: Сесії користувачів
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token TEXT NOT NULL UNIQUE CHECK(length(token) > 0),
    expires_at INTEGER NOT NULL CHECK(expires_at > 0),
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    ip_address TEXT,
    user_agent TEXT,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Індекси для швидкого пошуку
CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON sessions(created_at DESC);

-- ----------------------------------------------------------------------------
-- Тригер для автоматичного видалення закінчених сесій (опціонально)
-- Закоментовано, бо краще використовувати окремий cleanup job
-- ----------------------------------------------------------------------------
/*
CREATE TRIGGER IF NOT EXISTS trg_cleanup_expired_sessions
AFTER INSERT ON sessions
BEGIN
    DELETE FROM sessions WHERE expires_at < strftime('%s', 'now');
END;
*/

-- ----------------------------------------------------------------------------
-- View: Активні сесії
-- ----------------------------------------------------------------------------
CREATE VIEW IF NOT EXISTS v_active_sessions AS
SELECT 
    s.id AS session_id,
    s.user_id,
    u.username,
    u.email,
    u.first_name || ' ' || u.last_name AS user_full_name,
    s.token,
    s.expires_at,
    s.created_at,
    s.ip_address,
    s.user_agent,
    (s.expires_at - strftime('%s', 'now')) AS remaining_seconds
FROM sessions s
JOIN users u ON s.user_id = u.id
WHERE s.expires_at > strftime('%s', 'now')
  AND u.is_active = 1
  AND u.deleted_at IS NULL
ORDER BY s.created_at DESC;

-- ----------------------------------------------------------------------------
-- Реєстрація міграції
-- ----------------------------------------------------------------------------
INSERT OR IGNORE INTO schema_migrations (version, description) VALUES
('002', 'Sessions table for user authentication');

-- ============================================================================
-- КІНЕЦЬ МІГРАЦІЇ 002
-- ============================================================================