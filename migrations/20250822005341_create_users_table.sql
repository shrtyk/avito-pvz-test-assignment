-- +goose Up
-- +goose StatementBegin
CREATE TYPE user_roles AS ENUM('moderator', 'employee');

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash BYTEA NOT NULL,
  role user_roles NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS user_roles;

-- +goose StatementEnd
