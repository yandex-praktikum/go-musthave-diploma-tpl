BEGIN;

CREATE SCHEMA IF NOT EXISTS gophermart;

CREATE TABLE IF NOT EXISTS gophermart.users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    login varchar(32) NOT NULL,
    hash varchar(128) NOT NULL,
    balance bigint default 0,
    version integer NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS login_idx ON gophermart.users USING BTREE(login);

COMMENT ON TABLE gophermart.users IS 'Таблица пользователей сервиса';

COMMENT ON COLUMN gophermart.users.id IS 'ID пользователя';
COMMENT ON COLUMN gophermart.users.login IS 'Логин пользователя';
COMMENT ON COLUMN gophermart.users.hash IS 'Зашифрованный пароль пользователя';
COMMENT ON COLUMN gophermart.users.balance IS 'Баланс баллов пользователя';
COMMENT ON COLUMN gophermart.users.version IS 'Версия профиля';
COMMENT ON COLUMN gophermart.users.created_at IS 'Время создание пользователя';

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'order_status' AND typnamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'gophermart')) THEN
        CREATE TYPE gophermart.order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS gophermart.orders (
    id VARCHAR(256) NOT NULL,
    user_id UUID NOT NULL,
    status gophermart.order_status NOT NULL DEFAULT 'NEW',
    accrual bigint DEFAULT 0,
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE gophermart.orders ADD PRIMARY KEY(id);

COMMENT ON TABLE gophermart.orders IS 'Таблица заказов';

COMMENT ON COLUMN gophermart.orders.id IS 'Уникальный идентификатор заказов';
COMMENT ON COLUMN gophermart.orders.user_id IS 'ID пользователя';
COMMENT ON COLUMN gophermart.orders.status IS 'Статус заказа';
COMMENT ON COLUMN gophermart.orders.accrual IS 'Начисления в копейках';
COMMENT ON COLUMN gophermart.orders.uploaded_at IS 'Дата получения заказа';
COMMENT ON COLUMN gophermart.orders.updated_at IS 'Дата обновления заказа';

CREATE INDEX IF NOT EXISTS status_idx ON gophermart.orders USING btree(status);

CREATE TABLE IF NOT EXISTS gophermart.withdrawns (
    id VARCHAR(256) NOT NULL DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    order_id VARCHAR(256) NOT NULL,
    withdrawn bigint DEFAULT 0,
    processed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE gophermart.withdrawns ADD PRIMARY KEY(id);

COMMENT ON TABLE gophermart.withdrawns IS 'Таблица списания бонусных баллов';

COMMENT ON COLUMN gophermart.withdrawns.id IS 'Уникальный идентификатор операции';
COMMENT ON COLUMN gophermart.withdrawns.user_id IS 'ID пользователя';
COMMENT ON COLUMN gophermart.withdrawns.order_id IS 'Номер заказа';
COMMENT ON COLUMN gophermart.withdrawns.withdrawn IS 'Сумма списаний';
COMMENT ON COLUMN gophermart.withdrawns.processed_at IS 'Дата списания';

CREATE INDEX IF NOT EXISTS user_idx ON gophermart.withdrawns USING btree(user_id, order_id);

COMMIT;
