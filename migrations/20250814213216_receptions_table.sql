-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS receptions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  pvz_id UUID NOT NULL,
  status reception_statuses NOT NULL DEFAULT 'in_progress',
  CONSTRAINT fk_pvz_id FOREIGN KEY (pvz_id) REFERENCES pvzs (id) ON DELETE CASCADE
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS receptions;

-- +goose StatementEnd
