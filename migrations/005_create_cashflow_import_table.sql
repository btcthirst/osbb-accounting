CREATE TABLE IF NOT EXISTS imported_cashflow_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    import_batch_id INTEGER NOT NULL,
    date DATETIME NOT NULL,
    contractor_name TEXT NOT NULL,
    operation_type TEXT NOT NULL, -- 'debit' or 'credit'
    amount REAL NOT NULL,
    category_code TEXT,
    description TEXT,
    is_migrated BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(import_batch_id) REFERENCES import_batches(id) ON DELETE CASCADE
);

CREATE INDEX idx_imported_cashflow_records_batch_id ON imported_cashflow_records(import_batch_id);
