-- +goose Up
-- +goose StatementBegin
CREATE TYPE reception_statuses AS ENUM('in_progress', 'close');

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS reception_statuses;

-- +goose StatementEnd
