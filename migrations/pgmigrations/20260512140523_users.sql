-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login TEXT NOT NULL UNIQUE,
    hash_password TEXT NOT NULL,
    created_at  TIMESTAMPTZ     DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ     DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_login ON users(login);

CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    order_number VARCHAR(50) NOT NULL UNIQUE,
    status order_status NOT NULL DEFAULT 'NEW',
    is_registered BOOL NOT NULL DEFAULT FALSE,
    accrual NUMERIC(10,2),
    withdrawn NUMERIC(10,2),
    created_at  TIMESTAMPTZ     DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ     DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_amounts CHECK (
        (accrual IS NOT NULL AND withdrawn IS NULL) OR
        (accrual IS NULL AND withdrawn IS NOT NULL) OR
        (accrual IS NULL AND withdrawn IS NULL)
    )

    CONSTRAINT fk_orders_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_orders_order_user_id ON orders(order_number, user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_orders_order_user_id;
DROP TABLE IF EXISTS orders;

DROP TYPE order_status;

DROP INDEX IF EXISTS idx_users_login;
DROP TABLE IF EXISTS users;