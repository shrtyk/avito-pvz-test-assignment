-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_receptions_created_at ON receptions (created_at);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_receptions_created_at;

-- +goose StatementEnd
