-- Откат изменения типа accrual
ALTER TABLE orders ALTER COLUMN accrual TYPE INTEGER USING accrual::INTEGER;
