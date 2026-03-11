CREATE TABLE IF NOT EXISTS orders (
    id          BIGSERIAL    PRIMARY KEY,
    user_id     BIGINT       NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    number      VARCHAR(255) NOT NULL UNIQUE,
    status      VARCHAR(50)  NOT NULL,
    accrual     BIGINT       NULL,
    uploaded_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders (user_id);
CREATE INDEX IF NOT EXISTS idx_orders_uploaded_at ON orders (user_id, uploaded_at DESC);
