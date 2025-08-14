-- +goose Up
-- +goose StatementBegin
CREATE TYPE cities AS ENUM('Москва', 'Казань', 'Санкт-Петербург');

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS cities;

-- +goose StatementEnd
