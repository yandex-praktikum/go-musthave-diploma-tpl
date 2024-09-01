-- +goose Up
-- +goose StatementBegin
CREATE TABLE withdrawals (
    id UUID PRIMARY KEY,
    "order" VARCHAR(100) NOT NULL,
    sum FLOAT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL,
    user_id UUID NOT NULL,
    CONSTRAINT fk_user
        FOREIGN KEY(user_id) 
        REFERENCES users(id)
);
CREATE INDEX withdrawals_order_index ON withdrawals("order");
CREATE INDEX withdrawals_user_id_index ON withdrawals(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE withdrawals;
-- +goose StatementEnd