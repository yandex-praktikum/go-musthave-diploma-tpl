CREATE TABLE gophermart.orders (
    id SERIAL PRIMARY KEY,
    number BIGINT NOT NULL,
    status TEXT NOT NULL,
    accrual INTEGER,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id INTEGER REFERENCES gophermart.users(id)
);
