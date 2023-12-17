CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users
(
    id UUID NOT NULL DEFAULT gen_random_uuid() , 
    login text NOT NULL UNIQUE,
    passhash text NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT users_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS orders
(
    order_num text NOT NULL UNIQUE,
	user_id UUID NOT NULL REFERENCES users(id),
    status text NOT NULL,
    accrual double precision,
    uploaded_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT orders_pkey PRIMARY KEY (order_num)
);


CREATE TABLE IF NOT EXISTS balance
(
    user_id UUID NOT NULL REFERENCES users(id),
    current double precision  DEFAULT 0,
    withdrawn double precision  DEFAULT 0,
    CONSTRAINT balance_pkey PRIMARY KEY (user_id)
);

CREATE TABLE IF NOT EXISTS withdrawals
(
    id UUID NOT NULL DEFAULT gen_random_uuid() , 
    user_id UUID NOT NULL REFERENCES users(id),
    order_num text NOT NULL,
    sum double precision NOT NULL,
    processed_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT withdrawals_pkey PRIMARY KEY (id)
);
