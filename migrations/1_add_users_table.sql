-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    uuid UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
