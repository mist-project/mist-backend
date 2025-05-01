-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS appserver_permission(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    appserver_id UUID NOT NULL,
    appuser_id UUID NOT NULL,
    read_all BOOLEAN DEFAULT TRUE,
    write_all BOOLEAN DEFAULT FALSE,
    delete_all BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (appserver_id) REFERENCES appserver(id) ON DELETE CASCADE,
    FOREIGN KEY (appuser_id) REFERENCES appuser(id) ON DELETE CASCADE,

    CONSTRAINT appserver_permission_uk_appserver_appuser UNIQUE (appserver_id, appuser_id)
);

CREATE TABLE IF NOT EXISTS channel_permission(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    channel_id UUID NOT NULL,
    appserver_role_id UUID NOT NULL,
    read_all BOOLEAN DEFAULT TRUE,
    write_all BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (channel_id) REFERENCES channel(id) ON DELETE CASCADE,
    FOREIGN KEY (appserver_role_id) REFERENCES appserver_role(id) ON DELETE CASCADE,

    CONSTRAINT channel_permission_uk_channel_appserver_role UNIQUE (channel_id, appserver_role_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS appserver_permission;
DROP TABLE IF EXISTS channel_permission;
-- +goose StatementEnd
