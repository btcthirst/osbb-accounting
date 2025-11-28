-- ============================================================================
-- МІГРАЦІЯ 003: Таблиці для імпорту даних з XLSX
-- Дата: 2025-11-28
-- Опис: Створення таблиць для зберігання імпортованих історичних даних
--       з можливістю подальшої міграції в операційні таблиці
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Таблиця: Батчі імпорту (журнал операцій імпорту)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS import_batches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_name TEXT NOT NULL CHECK(length(file_name) >= 1),
    file_path TEXT,
    
    -- Статус імпорту
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'processing', 'completed', 'failed')),
    
    -- Статистика
    total_sheets INTEGER DEFAULT 0,
    total_records INTEGER DEFAULT 0,
    successful_records INTEGER DEFAULT 0,
    failed_records INTEGER DEFAULT 0,
    
    -- Помилки
    error_message TEXT,
    
    -- Метадані (JSON)
    metadata TEXT, -- JSON з додатковою інформацією про файл
    
    -- Аудит
    imported_by INTEGER NOT NULL,
    imported_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    completed_at INTEGER,
    
    -- Timestamps
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    
    FOREIGN KEY (imported_by) REFERENCES users(id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_import_batches_status ON import_batches(status);
CREATE INDEX IF NOT EXISTS idx_import_batches_imported_by ON import_batches(imported_by);
CREATE INDEX IF NOT EXISTS idx_import_batches_imported_at ON import_batches(imported_at DESC);

-- ----------------------------------------------------------------------------
-- Таблиця: Імпортовані місячні записи (історичні дані з XLSX)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS imported_monthly_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    import_batch_id INTEGER NOT NULL,
    
    -- Період
    period_month INTEGER NOT NULL CHECK(period_month BETWEEN 1 AND 12),
    period_year INTEGER NOT NULL CHECK(period_year >= 2020 AND period_year <= 2100),
    sheet_name TEXT, -- Назва аркушу (січень, лютий, тощо)
    
    -- Дані з XLSX (відповідно до колонок)
    apartment_number TEXT NOT NULL CHECK(length(apartment_number) >= 1),
    owner_name TEXT NOT NULL CHECK(length(owner_name) >= 1),
    account_number TEXT, -- Особовий рахунок
    
    -- Баланси (Д-Т / К-Т)
    opening_debit REAL DEFAULT 0 CHECK(opening_debit >= 0),
    opening_credit REAL DEFAULT 0 CHECK(opening_credit >= 0),
    closing_debit REAL DEFAULT 0 CHECK(closing_debit >= 0),
    closing_credit REAL DEFAULT 0 CHECK(closing_credit >= 0),
    
    -- Інформація про квартиру
    total_area REAL CHECK(total_area IS NULL OR total_area > 0),
    discount_area REAL DEFAULT 0 CHECK(discount_area >= 0),
    discount_percent INTEGER DEFAULT 0 CHECK(discount_percent >= 0 AND discount_percent <= 100),
    
    -- Нарахування
    tariff REAL CHECK(tariff IS NULL OR tariff > 0),
    charge_amount REAL DEFAULT 0 CHECK(charge_amount >= 0), -- 100% нарах.
    discount_amount REAL DEFAULT 0 CHECK(discount_amount >= 0),
    amount_due REAL DEFAULT 0 CHECK(amount_due >= 0), -- До сплати
    
    -- Платежі
    amount_paid REAL DEFAULT 0 CHECK(amount_paid >= 0), -- Сплачено
    
    -- Інше
    corrections REAL DEFAULT 0, -- Коригування (може бути + або -)
    notes TEXT,
    
    -- Зв'язки з операційними таблицями (NULL = не зв'язано)
    apartment_id INTEGER,
    owner_id INTEGER,
    ownership_share_id INTEGER,
    charge_id INTEGER,
    payment_id INTEGER,
    
    -- Статус міграції
    is_migrated INTEGER NOT NULL DEFAULT 0 CHECK(is_migrated IN (0, 1)),
    migrated_at INTEGER,
    
    -- Timestamps
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    
    FOREIGN KEY (import_batch_id) REFERENCES import_batches(id) ON DELETE CASCADE,
    FOREIGN KEY (apartment_id) REFERENCES apartments(id) ON DELETE SET NULL,
    FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE SET NULL,
    FOREIGN KEY (ownership_share_id) REFERENCES ownership_shares(id) ON DELETE SET NULL,
    FOREIGN KEY (charge_id) REFERENCES charges(id) ON DELETE SET NULL,
    FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL
);

-- Індекси для швидкого пошуку
CREATE INDEX IF NOT EXISTS idx_imported_records_batch ON imported_monthly_records(import_batch_id);
CREATE INDEX IF NOT EXISTS idx_imported_records_period ON imported_monthly_records(period_year DESC, period_month DESC);
CREATE INDEX IF NOT EXISTS idx_imported_records_apartment ON imported_monthly_records(apartment_number);
CREATE INDEX IF NOT EXISTS idx_imported_records_owner ON imported_monthly_records(owner_name);
CREATE INDEX IF NOT EXISTS idx_imported_records_migrated ON imported_monthly_records(is_migrated);
CREATE INDEX IF NOT EXISTS idx_imported_records_operational_links ON imported_monthly_records(apartment_id, owner_id, ownership_share_id);

-- ----------------------------------------------------------------------------
-- Тригер: Автоматичне оновлення updated_at
-- ----------------------------------------------------------------------------
CREATE TRIGGER IF NOT EXISTS trg_import_batches_updated_at 
AFTER UPDATE ON import_batches
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE import_batches SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS trg_imported_records_updated_at 
AFTER UPDATE ON imported_monthly_records
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE imported_monthly_records SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
END;

-- ============================================================================
-- Коментарі та документація
-- ============================================================================

/*
ПРИЗНАЧЕННЯ ТАБЛИЦЬ:

1. import_batches:
   - Зберігає інформацію про кожну операцію імпорту
   - Дозволяє відстежувати історію імпортів
   - Містить статистику та можливі помилки

2. imported_monthly_records:
   - Зберігає детальні дані з кожного місячного аркушу XLSX
   - Зберігає ВСІ колонки з файлу для повного відновлення
   - Опціонально зв'язується з операційними таблицями
   - Поле is_migrated показує чи були дані перенесені

РОБОЧИЙ ПРОЦЕС:
1. Імпорт → створюється batch, парсяться дані, записуються records
2. Перегляд → користувач може бачити імпортовані дані
3. Міграція (опціонально) → створюємо apartments, owners, charges, payments
   та зв'язуємо через apartment_id, owner_id тощо
*/
