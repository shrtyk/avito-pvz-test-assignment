-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_products_reception_id ON products (reception_id);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_products_reception_id;

-- +goose StatementEnd
