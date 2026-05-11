-- Изменяем тип accrual на DOUBLE PRECISION для сохранения дробных значений
ALTER TABLE orders ALTER COLUMN accrual TYPE DOUBLE PRECISION USING accrual::DOUBLE PRECISION;
