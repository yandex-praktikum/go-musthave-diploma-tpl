CREATE TABLE IF NOT EXISTS withdrawals (
    id           BIGSERIAL    PRIMARY KEY,
    user_id      BIGINT       NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    "order"      VARCHAR(255) NOT NULL,
    sum          BIGINT       NOT NULL CHECK (sum > 0),
    processed_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (user_id, "order")
);

CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id ON withdrawals (user_id);
CREATE INDEX IF NOT EXISTS idx_withdrawals_user_processed_at ON withdrawals (user_id, processed_at DESC);
