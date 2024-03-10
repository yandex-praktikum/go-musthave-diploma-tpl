-- +goose Up
-- +goose StatementBegin
SELECT 'goose up init';

SET lock_timeout TO '10s';
SET statement_timeout to '20s';

CREATE SCHEMA IF NOT EXISTS public;
CREATE SCHEMA IF NOT EXISTS gophermart;

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
SELECT 'goose up init';
-- +goose StatementEnd
