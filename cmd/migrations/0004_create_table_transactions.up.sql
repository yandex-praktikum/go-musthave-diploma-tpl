-- Создание таблицы
CREATE table IF NOT EXISTS t_gophermart.t_transactions  (
    n_id SERIAL PRIMARY KEY,
    n_order INT,
    s_user VARCHAR(100) NOT NULL,
    s_type VARCHAR(10) NOT NULL CHECK (s_type IN ('plus', 'minus')),
    n_value DECIMAL(10, 2) NOT NULL,
    dt_created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (n_order) REFERENCES t_gophermart.t_orders(n_order) ON DELETE SET NULL,
    FOREIGN KEY (s_user) REFERENCES t_gophermart.t_users(s_login) ON DELETE CASCADE
);

COMMENT ON TABLE  t_gophermart.t_transactions IS 'Таблица ссылок';
COMMENT ON COLUMN t_gophermart.t_transactions.n_id IS 'ид транзакции';
COMMENT ON COLUMN t_gophermart.t_transactions.n_order IS 'ид заказа';
COMMENT ON COLUMN t_gophermart.t_transactions.s_type IS 'тип - начисление/списание';
COMMENT ON COLUMN t_gophermart.t_transactions.n_value IS 'количество начислений/списаний';
COMMENT ON COLUMN t_gophermart.t_transactions.s_user IS 'ползователь';
COMMENT ON COLUMN t_gophermart.t_transactions.dt_created_at IS 'дата списания бонусов';
