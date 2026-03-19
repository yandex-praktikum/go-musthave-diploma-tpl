-- +goose Up
-- +goose StatementBegin

-- 1. Таблица балансов (агрегат накопительных итогов)
CREATE TABLE IF NOT EXISTS balances (
    user_id     BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    accrual     BIGINT NOT NULL DEFAULT 0,  -- Сумма всех начислений
    withdrawn   BIGINT NOT NULL DEFAULT 0,  -- Сумма всех списаний
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    
    -- Главная бизнес-проверка: баланс не может быть отрицательным
    CONSTRAINT check_balance_not_negative CHECK (accrual - withdrawn >= 0)
);

-- 2. Таблица списаний
CREATE TABLE IF NOT EXISTS withdrawals (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_number  VARCHAR(255) NOT NULL UNIQUE,
    amount        BIGINT NOT NULL, 
    processed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Функция синхронизации начислений (из orders)
CREATE OR REPLACE FUNCTION fn_sync_accrual_to_balance()
RETURNS TRIGGER AS $$
DECLARE
    delta BIGINT := 0;
    uid   BIGINT;
BEGIN
    IF (TG_OP = 'INSERT') THEN
        delta := COALESCE(NEW.accrual, 0);
        uid   := NEW.user_id;
    ELSIF (TG_OP = 'UPDATE') THEN
        delta := COALESCE(NEW.accrual, 0) - COALESCE(OLD.accrual, 0);
        uid   := NEW.user_id;
    ELSIF (TG_OP = 'DELETE') THEN
        delta := -COALESCE(OLD.accrual, 0);
        uid   := OLD.user_id;
    END IF;

    IF delta <> 0 THEN
        INSERT INTO balances (user_id, accrual, withdrawn)
        VALUES (uid, delta, 0)
        ON CONFLICT (user_id) DO UPDATE
        SET accrual = balances.accrual + delta,
            updated_at = NOW();
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 4. Функция синхронизации списаний (из withdrawals)
CREATE OR REPLACE FUNCTION fn_sync_withdrawn_to_balance()
RETURNS TRIGGER AS $$
DECLARE
    delta BIGINT := 0;
    uid   BIGINT;
BEGIN
    IF (TG_OP = 'INSERT') THEN
        delta := NEW.amount;
        uid   := NEW.user_id;
    ELSIF (TG_OP = 'UPDATE') THEN
        delta := NEW.amount - OLD.amount;
        uid   := NEW.user_id;
    ELSIF (TG_OP = 'DELETE') THEN
        delta := -OLD.amount;
        uid   := OLD.user_id;
    END IF;

    IF delta <> 0 THEN
        INSERT INTO balances (user_id, accrual, withdrawn)
        VALUES (uid, 0, delta)
        ON CONFLICT (user_id) DO UPDATE
        SET withdrawn = balances.withdrawn + delta,
            updated_at = NOW();
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 5. Регистрация триггеров
CREATE TRIGGER trg_orders_to_balance 
AFTER INSERT OR UPDATE OR DELETE ON orders 
FOR EACH ROW EXECUTE FUNCTION fn_sync_accrual_to_balance();

CREATE TRIGGER trg_withdrawals_to_balance 
AFTER INSERT OR UPDATE OR DELETE ON withdrawals 
FOR EACH ROW EXECUTE FUNCTION fn_sync_withdrawn_to_balance();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_withdrawals_to_balance ON withdrawals;
DROP TRIGGER IF EXISTS trg_orders_to_balance ON orders;
DROP FUNCTION IF EXISTS fn_sync_withdrawn_to_balance();
DROP FUNCTION IF EXISTS fn_sync_accrual_to_balance();
DROP TABLE IF EXISTS withdrawals;
DROP TABLE IF EXISTS balances;
-- +goose StatementEnd
