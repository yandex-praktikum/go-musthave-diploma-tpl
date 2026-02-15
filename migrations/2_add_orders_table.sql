-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders (
    number TEXT PRIMARY KEY,
    user_uuid UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'NEW',
    accrual NUMERIC(10,2),
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),

    CONSTRAINT fk_orders_user
        FOREIGN KEY (user_uuid)
        REFERENCES users (uuid)
        ON DELETE CASCADE
);

CREATE INDEX idx_orders_user_uuid ON orders(user_uuid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE orders;
-- +goose StatementEnd
