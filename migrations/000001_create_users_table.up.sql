CREATE TABLE IF NOT EXISTS users (
    user_id   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username  VARCHAR(255) NOT NULL UNIQUE,
    password  VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS accruals_journal (
    order_id        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(user_id),
    order_number    TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'NEW' CHECK (status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED')),
    amount          NUMERIC(10, 2) CHECK (status != 'PROCESSED' OR amount IS NOT NULL),
    uploaded_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS withdrawals_journal (
    withdrawal_id   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_number    TEXT NOT NULL UNIQUE,
    user_id         BIGINT NOT NULL REFERENCES users(user_id),
    amount          NUMERIC(10, 2) NOT NULL CHECK (amount > 0), 
    processed_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);