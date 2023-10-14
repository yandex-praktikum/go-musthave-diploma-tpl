CREATE TABLE IF NOT EXISTS orders
(
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER REFERENCES users (id) NOT NULL,
    order_id    VARCHAR(20)                   NOT NULL,
    status      VARCHAR(20)                   NOT NULL,
    accrual     NUMERIC(10) DEFAULT 0,
    withdraw    NUMERIC(10) DEFAULT 0,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (order_id)
);