-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_receptions_pvz_id ON receptions (pvz_id);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_receptions_pvz_id;

-- +goose StatementEnd
