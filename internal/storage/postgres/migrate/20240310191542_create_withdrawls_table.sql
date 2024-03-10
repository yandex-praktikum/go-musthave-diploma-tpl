-- +goose Up
-- +goose StatementBegin
SELECT 'goose up create_withdrawals_table';

SET lock_timeout TO '10s';
SET statement_timeout to '20s';

CREATE TABLE IF NOT EXISTS gophermart.withdrawals
(
    id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid        NOT NULL REFERENCES gophermart.users(id),
    amount     bigint      NOT NULL,
    created_at timestamptz NOT NULL DEFAULT current_timestamp
);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
SELECT 'goose down create_withdrawals_table';

DROP TABLE IF EXISTS gophermart.withdrawals;

-- +goose StatementEnd
