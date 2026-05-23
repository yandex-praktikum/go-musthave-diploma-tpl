-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    login VARCHAR(255) NOT NULL UNIQUE,
    password_hash BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индекс для быстрого поиска по логину
CREATE INDEX IF NOT EXISTS idx_users_login ON users(login);

-- Таблица заказов
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    number VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'NEW',
    accrual INTEGER,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Уникальность номера заказа в рамках пользователя
CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_user_number ON orders(user_id, number);
-- Индекс для проверки глобальной уникальности номера
CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_number ON orders(number);
-- Индекс для быстрой выборки заказов пользователя
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);

-- Таблица транзакций баланса (начисления и списания)
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- 'charge' или 'withdraw'
    amount NUMERIC(10, 2) NOT NULL,
    order_number VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индекс для выборки транзакций пользователя
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
-- Индекс для выборки по номеру заказа
CREATE INDEX IF NOT EXISTS idx_transactions_order_number ON transactions(order_number);

-- Таблица списаний (withdrawals)
CREATE TABLE IF NOT EXISTS withdrawals (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_number VARCHAR(255) NOT NULL,
    sum NUMERIC(10, 2) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индекс для выборки списаний пользователя
CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id ON withdrawals(user_id);
-- Индекс для сортировки по времени
CREATE INDEX IF NOT EXISTS idx_withdrawals_processed_at ON withdrawals(processed_at DESC);
