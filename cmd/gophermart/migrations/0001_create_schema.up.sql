-- +goose Up
-- Создание схемы
CREATE SCHEMA IF NOT EXISTS t_gophermart;

-- +goose Down
DROP SCHEMA IF EXISTS t_gophermart CASCADE;