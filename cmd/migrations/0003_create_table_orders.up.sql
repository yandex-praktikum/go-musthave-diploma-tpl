-- +goose Up
-- Создание таблицы
CREATE table IF NOT EXISTS t_gophermart.t_orders (
    n_order bigint PRIMARY KEY,
    s_user VARCHAR(100) NOT NULL,
    s_status VARCHAR(50) NOT NULL,
    s_sber_thx VARCHAR(50) NOT NULL,
    dt_created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (s_user) REFERENCES t_gophermart.t_users(s_login) ON DELETE CASCADE
);

COMMENT ON TABLE  t_gophermart.t_orders IS 'Таблица ссылок';
COMMENT ON COLUMN t_gophermart.t_orders.n_order IS 'идентификатор заказа';
COMMENT ON COLUMN t_gophermart.t_orders.s_user IS 'пользователь';
COMMENT ON COLUMN t_gophermart.t_orders.s_status IS 'статус заказа';
COMMENT ON COLUMN t_gophermart.t_orders.s_sber_thx IS 'количество кешбека за заказ';
COMMENT ON COLUMN t_gophermart.t_orders.dt_created_at IS 'дата регистрации заказа в сервисе';
-- +goose Down
DROP TABLE IF EXISTS t_gophermart.t_orders CASCADE;