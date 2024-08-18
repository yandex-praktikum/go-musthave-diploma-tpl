-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id UUID PRIMARY KEY,
    name VARCHAR(255),
    age SMALLINT,
    email VARCHAR(255) NOT NULL UNIQUE,
    email_confirmed BOOLEAN,
    balance INTEGER,
    withdrawn INTEGER,
    password VARCHAR(255) NOT NULL,
    refresh_token VARCHAR(255)
);
CREATE INDEX users_email_index ON users (email);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
