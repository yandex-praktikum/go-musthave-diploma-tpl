CREATE TABLE gophermart.withdrawals (
    id SERIAL PRIMARY KEY,
    "order" TEXT NOT NULL,
    sum DECIMAL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id INTEGER REFERENCES gophermart.users(id)
);
