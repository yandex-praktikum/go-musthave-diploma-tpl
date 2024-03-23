-- +goose Up
-- +goose StatementBegin
SELECT 'goose up create_orders_table';

SET lock_timeout TO '10s';
SET statement_timeout to '20s';

CREATE OR REPLACE FUNCTION create_update_timestamp()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at := current_timestamp;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS gophermart.orders
(
    id         bigint        PRIMARY KEY,
    user_id    uuid        NOT NULL REFERENCES gophermart.users(id),
    status     text        NOT NULL DEFAULT 'NEW',
    accrual    bigint      NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at timestamptz NOT NULL DEFAULT current_timestamp
);

CREATE TRIGGER create_update_timestamp
    BEFORE UPDATE ON gophermart.orders
    FOR EACH ROW EXECUTE PROCEDURE create_update_timestamp();

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
SELECT 'goose down create_orders_table';

DROP TRIGGER create_update_timestamp on gophermart.orders;
DROP TABLE IF EXISTS gophermart.orders;
DROP FUNCTION IF EXISTS create_update_timestamp();

-- +goose StatementEnd
