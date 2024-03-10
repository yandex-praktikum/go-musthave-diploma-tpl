-- +goose Up
-- +goose StatementBegin
SELECT 'goose up create_orders_table';

SET lock_timeout TO '10s';
SET statement_timeout to '20s';

CREATE TABLE IF NOT EXISTS gophermart.orders
(
    id         bigint        PRIMARY KEY,
    user_id    uuid        NOT NULL REFERENCES gophermart.users(id),
    status     text        NOT NULL,
    accrual    bigint      NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at timestamptz NOT NULL DEFAULT current_timestamp
);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
SELECT 'goose down create_orders_table';

DROP TABLE IF EXISTS gophermart.orders;

-- +goose StatementEnd
