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
