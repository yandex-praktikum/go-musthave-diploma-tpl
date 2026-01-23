-- +goose Up
-- Создание таблицы

CREATE table IF NOT EXISTS t_gophermart.t_users (
    s_login VARCHAR(100) PRIMARY KEY,
    s_pass_hash VARCHAR(200)
);

COMMENT ON TABLE  t_gophermart.t_users IS 'Таблица пользователей';
COMMENT ON COLUMN t_gophermart.t_users.s_login IS 'Логин ';
COMMENT ON COLUMN t_gophermart.t_users.s_pass_hash IS 'Хэш пароля';
-- +goose Down
DROP TABLE IF EXISTS t_gophermart.t_users CASCADE;