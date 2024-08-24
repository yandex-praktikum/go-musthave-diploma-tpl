-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id UUID PRIMARY KEY,
    name VARCHAR(255),
    age SMALLINT,
    username VARCHAR(255) NOT NULL,
    balance FLOAT NOT NULL,
    withdrawn FLOAT NOT NULL,
    password VARCHAR(255) NOT NULL,
    refresh_token VARCHAR(255)
);
CREATE INDEX users_username_index ON users (username);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
