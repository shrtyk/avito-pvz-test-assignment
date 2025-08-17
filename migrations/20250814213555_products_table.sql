-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS products (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  reception_id UUID NOT NULL,
  type product_types NOT NULL,
  CONSTRAINT fk_reception_id FOREIGN KEY (reception_id) REFERENCES receptions (id) ON DELETE CASCADE
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS products;

-- +goose StatementEnd
