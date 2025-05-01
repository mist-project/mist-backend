-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS channel_role (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    appserver_id UUID NOT NULL,
    channel_id UUID NOT NULL,
    appserver_role_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (appserver_role_id) REFERENCES appserver_role(id) ON DELETE CASCADE,
    FOREIGN KEY (appserver_id, channel_id) REFERENCES channel(appserver_id, id) ON DELETE CASCADE,
    CONSTRAINT channel_role_uk_role_channel UNIQUE (channel_id, appserver_role_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS channel_role;
-- +goose StatementEnd
