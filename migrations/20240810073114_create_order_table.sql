-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    number VARCHAR(100) NOT NULL UNIQUE,
    accrual FLOAT  NOT NULL,
    status VARCHAR(50) NOT NULL,
    user_id UUID NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT fk_user
        FOREIGN KEY(user_id) 
        REFERENCES users(id)
);
CREATE INDEX orders_number_index ON orders(number);
CREATE INDEX orders_user_id_index ON orders(user_id);
CREATE INDEX orders_uploaded_at_index ON orders(uploaded_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE orders;
-- +goose StatementEnd
