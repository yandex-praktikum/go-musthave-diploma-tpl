-- +goose Up
CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE IF NOT EXISTS users
(
    id         SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    login      VARCHAR(255) NOT NULL,
    password   VARCHAR(255) NOT NULL,
    is_active  BOOLEAN      NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS orders
(
    id         SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    number     VARCHAR(255) NOT NULL,
    status     order_status NOT NULL DEFAULT 'NEW',
    accrual    DECIMAL(10, 2),
    user_id    INTEGER      NOT NULL REFERENCES users (id) ON DELETE CASCADE
);

-- Индексы для быстрого поиска
CREATE INDEX idx_orders_user_id ON orders (user_id);
CREATE UNIQUE INDEX idx_users_login ON users (login);
CREATE UNIQUE INDEX idx_orders_number ON orders (number);

-- +goose Down
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS order_status;