-- +goose Up
-- +goose StatementBegin
    CREATE TABLE IF NOT EXISTS orders (
        order_number  VARCHAR(255) NOT NULL PRIMARY KEY,                 -- Номер заказа - строка чисел
        user_id       BIGINT NOT NULL ,                                  -- user ID
        status        VARCHAR(20) NOT NULL DEFAULT 'NEW',                -- Статус обработки заказа
        uploaded_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,    -- Дата загрузки заказа
        accrual       BIGINT NOT NULL DEFAULT 0,                         -- Поле для баллов в копейках (из ТЗ)
        accrualed_at  TIMESTAMPTZ DEFAULT NULL,                          -- Дата последнего обращения к accrual сервису

        CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
        CONSTRAINT check_status CHECK (status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED'))
    );

    -- Уникальный индекс для номера заказа.
    -- INCLUDE позволяет UNION-запросу отработать через Index Only Scan.
    CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_order_number_inc
    ON orders (order_number) INCLUDE (user_id);

    -- Составной индекс для метода List.
    -- Позволяет мгновенно найти заказы юзера и вернуть их уже отсортированными.
    CREATE INDEX IF NOT EXISTS idx_orders_user_uploaded
    ON orders (user_id, uploaded_at DESC);

    -- Partial Index только для активных задач.
    -- Позволяет воркерам мгновенно находить еще не обработанные ордеры
    CREATE INDEX IF NOT EXISTS idx_orders_worker_queue
    ON orders (uploaded_at ASC, accrualed_at)
    WHERE status IN ('NEW', 'PROCESSING');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
    DROP TABLE IF EXISTS orders;
-- +goose StatementEnd