-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    login         VARCHAR(255) NOT NULL UNIQUE,
    password_hash CHAR(60) NOT NULL,
    created_at    TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DO $$ 
BEGIN 
    -- Удаление таблицы пользователей запрещено для предотвращения потери данных.
    NULL; 
END $$;

-- +goose StatementEnd
