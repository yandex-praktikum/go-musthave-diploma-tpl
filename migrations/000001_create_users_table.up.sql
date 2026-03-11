-- Таблица пользователей (регистрация: login + password). Пароль храним как хэш;
CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL       PRIMARY KEY,
    login       VARCHAR(255)    NOT NULL UNIQUE,
    password_hash VARCHAR(255)  NOT NULL,
    active      BOOLEAN         NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_users_login ON users (login);
