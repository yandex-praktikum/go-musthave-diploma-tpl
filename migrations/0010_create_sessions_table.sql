-- +goose Up
-- +goose StatementBegin
    CREATE TABLE IF NOT EXISTS sessions (
        id            UUID PRIMARY KEY DEFAULT gen_random_uuid(), 
        user_id       BIGINT NOT NULL,
        
        user_agent    TEXT,
        ip_address    VARCHAR(255),
        device_id     UUID DEFAULT NULL,
        
        expires_at    TIMESTAMPTZ NOT NULL,
        created_at    TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
        updated_at    TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

        CONSTRAINT fk_session_user 
            FOREIGN KEY (user_id) 
            REFERENCES users(id) 
            ON DELETE CASCADE
    );

    -- Индекс для быстрой очистки просроченных сессий и поиска по юзеру
    CREATE INDEX IF NOT EXISTS idx_sessions_user_expires 
    ON sessions (user_id, expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
    DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd