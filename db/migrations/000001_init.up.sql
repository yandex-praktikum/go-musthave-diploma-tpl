CREATE TABLE IF NOT EXISTS users(
    id uuid PRIMARY KEY,
    login varchar UNIQUE NOT NULL,
    password varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS orders(
    id uuid PRIMARY KEY,
    order_number varchar UNIQUE NOT NULL,
    user_id uuid NOT NULL REFERENCES user(id) ON DELETE CASCADE,
    status varchar NOT NULL,
    accrual numeric(11, 3) NOT NULL,
    uploaded_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS withdrawls(
    id uuid PRIMARY KEY,
    order_number varchar UNIQUE NOT NULL,
    user_id uuid NOT NULL REFERENCES user(id) ON DELETE CASCADE,
    amount numeric(11, 3) NOT NULL,
    status varchar DEFAULT 'FAILURE',
    processed_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS balance(
    user_id PRIMARY KEY REFERENCES user(id) ON DELETE CASCADE,
    balance numeric(11, 3) NOT NULL DEFAULT 0,
    uploaded_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS withdrawl_balances(
    user_id uuid PRIMARY KEY,
    amount numeric(11, 3) DEFAULT 0,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP
);