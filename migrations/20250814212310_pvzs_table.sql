-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS pvzs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  city cities NOT NULL
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pvzs;

-- +goose StatementEnd
