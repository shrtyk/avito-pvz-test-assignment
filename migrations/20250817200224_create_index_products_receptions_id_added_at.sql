-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_products_reception_id_added_at ON products (reception_id, added_at DESC);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_products_reception_id_added_at;

-- +goose StatementEnd
