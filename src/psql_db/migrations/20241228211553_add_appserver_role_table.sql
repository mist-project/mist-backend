-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS appserver_role (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    appserver_id UUID NOT NULL,
    name VARCHAR(64) NOT NULL,

    appserver_permission_mask BIGINT NOT NULL DEFAULT 0, 
    channel_permission_mask BIGINT NOT NULL DEFAULT 0,
    sub_permission_mask BIGINT NOT NULL DEFAULT 0,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (appserver_id) REFERENCES appserver(id) ON DELETE CASCADE,
    CONSTRAINT appserver_role_uk_appserver_name UNIQUE (appserver_id, name),
    CONSTRAINT appserver_role_uk_server_role UNIQUE (appserver_id, id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS appserver_role;
-- +goose StatementEnd
