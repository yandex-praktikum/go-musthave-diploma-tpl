-- +goose Up
-- +goose StatementBegin
SELECT 'goose up create_users_table';

SET lock_timeout TO '10s';
SET statement_timeout to '20s';

CREATE TABLE IF NOT EXISTS gophermart.users
(
    id        uuid UNIQUE PRIMARY KEY DEFAULT gen_random_uuid(),
    login     text UNIQUE NOT NULL,
    pass_hash bytea NOT NULL
);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
SELECT 'goose down create_users_table';

DROP TABLE IF EXISTS gophermart.users;

-- +goose StatementEnd
