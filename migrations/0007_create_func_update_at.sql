-- +goose Up
-- +goose StatementBegin

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Удаляем только если уверены, что другие таблицы её не используют
DROP FUNCTION IF EXISTS update_updated_at_column();

-- +goose StatementEnd
