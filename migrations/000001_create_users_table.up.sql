CREATE TABLE users (
    user_id   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username  VARCHAR(255) NOT NULL UNIQUE,
    password  VARCHAR(255) NOT NULL
);

CREATE TABLE orders (
    order_id        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(user_id),
    order_number    TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'NEW' CHECK (status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED')),
    accrual         NUMERIC(10, 2),
    uploaded_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);