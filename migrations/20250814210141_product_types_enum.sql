-- +goose Up
-- +goose StatementBegin
CREATE TYPE product_types AS ENUM('одежда', 'электроника', 'обувь');

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS product_types;

-- +goose StatementEnd
