-- ============================================================================
-- МІГРАЦІЯ 004: Платежі контрагентів та стандартні категорії витрат
-- Сумісність: SQLite 3.35+
-- Дата: 2025-11-29
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Таблиця: Платежі від контрагентів (орендна плата, утримання)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS contractor_payments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    contractor_id INTEGER NOT NULL,
    payment_date INTEGER NOT NULL,
    amount REAL NOT NULL CHECK(amount > 0),
    payment_method TEXT NOT NULL CHECK(payment_method IN ('cash', 'card', 'bank_transfer', 'other')),
    purpose TEXT NOT NULL CHECK(length(purpose) >= 3),
    
    -- Період, за який здійснюється платіж (опціонально)
    period_month INTEGER CHECK(period_month IS NULL OR period_month BETWEEN 1 AND 12),
    period_year INTEGER CHECK(period_year IS NULL OR (period_year >= 2020 AND period_year <= 2100)),
    
    receipt_number TEXT,
    notes TEXT,
    
    -- Soft Delete
    deleted_at INTEGER DEFAULT NULL,
    
    -- Timestamps
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    
    FOREIGN KEY (contractor_id) REFERENCES contractors(id) ON DELETE RESTRICT
);

-- Індекси
CREATE INDEX IF NOT EXISTS idx_contractor_payments_contractor ON contractor_payments(contractor_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractor_payments_date ON contractor_payments(payment_date DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractor_payments_period ON contractor_payments(period_year DESC, period_month DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_contractor_payments_deleted_at ON contractor_payments(deleted_at) WHERE deleted_at IS NOT NULL;

-- Тригер для автоматичного оновлення updated_at
CREATE TRIGGER IF NOT EXISTS trg_contractor_payments_updated_at 
AFTER UPDATE ON contractor_payments
FOR EACH ROW
WHEN NEW.updated_at = OLD.updated_at AND NEW.deleted_at IS NULL
BEGIN
    UPDATE contractor_payments SET updated_at = strftime('%s', 'now') WHERE id = NEW.id;
END;

-- ----------------------------------------------------------------------------
-- Стандартні категорії витрат для Cash Flow
-- ----------------------------------------------------------------------------
INSERT OR IGNORE INTO expense_categories (name, code, category_type, description, is_active) VALUES
('Кошти на картку', '313', 'other', 'Зняття готівки / переказ на картку', 1),
('Електроенергія', '63', 'utility', 'Оплата за електроенергію', 1),
('ПДФО 18%', '641', 'salary', 'Податок на доходи фізичних осіб 18%', 1),
('Військовий збір', '641.1', 'salary', 'Військовий збір 1.5%', 1),
('ЄСВ 22%', '651', 'salary', 'Єдиний соціальний внесок 22%', 1),
('Комісія банку', '94', 'service', 'Банківська комісія за обслуговування', 1);
