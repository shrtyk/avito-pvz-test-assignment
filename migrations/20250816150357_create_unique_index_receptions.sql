-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX one_in_progress_reception_per_pvz_id ON receptions (pvz_id, status)
WHERE
  status = 'in_progress';

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS one_in_progress_reception_per_pvz_id;

-- +goose StatementEnd
