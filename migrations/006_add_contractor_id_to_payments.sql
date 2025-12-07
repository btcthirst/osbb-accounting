ALTER TABLE payments ADD COLUMN contractor_id INTEGER REFERENCES contractors(id);
CREATE INDEX idx_payments_contractor_id ON payments(contractor_id);
