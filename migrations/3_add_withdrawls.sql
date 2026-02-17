-- +goose Up
-- +goose StatementBegin
CREATE TABLE withdrawals (
    "order" TEXT PRIMARY KEY,
    user_uuid UUID NOT NULL,
    sum NUMERIC(10,2) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),

    CONSTRAINT fk_withdrawals_user
        FOREIGN KEY (user_uuid)
        REFERENCES users (uuid)
        ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE withdrawals;
-- +goose StatementEnd
