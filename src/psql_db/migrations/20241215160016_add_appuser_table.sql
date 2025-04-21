-- +goose Up
-- +goose StatementBegin
-- Create appuser status enum
CREATE TYPE appuser_online_status AS ENUM (
    'inactive',
    'online',
    'offline',
    'away'
);

-- Create appuser table
CREATE TABLE IF NOT EXISTS appuser (
    id UUID PRIMARY KEY NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    online_status appuser_online_status NOT NULL DEFAULT 'offline',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS appuser;
DROP TYPE IF EXISTS appuser_online_status;
-- +goose StatementEnd
