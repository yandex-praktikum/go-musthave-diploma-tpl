CREATE TABLE gophermart.orders (
    id SERIAL PRIMARY KEY,
    number TEXT NOT NULL,
-- - `REGISTERED` — заказ зарегистрирован, но вознаграждение не рассчитано;
-- - `INVALID` — заказ не принят к расчёту, и вознаграждение не будет начислено;
-- - `PROCESSING` — расчёт начисления в процессе;
-- - `PROCESSED` — расчёт начисления окончен;
    status TEXT NOT NULL,
    accrual NUMERIC,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id INTEGER REFERENCES gophermart.users(id)
);
