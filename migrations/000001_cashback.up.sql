CREATE TABLE personal_account (
    id SERIAL PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    score NUMERIC,
    total NUMERIC
);
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    order_number TEXT NOT NULL UNIQUE,
    user_id INTEGER NOT NULL REFERENCES personal_account(id),
    status TEXT,
    accrual NUMERIC,
    uploaded_at TIMESTAMP DEFAULT now()
);