-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS appserver_role_sub (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    appuser_id UUID NOT NULL,
    appserver_sub_id UUID NOT NULL,
    appserver_role_id UUID NOT NULL,
    appserver_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (appserver_role_id) REFERENCES appserver_role(id) ON DELETE CASCADE,
    FOREIGN KEY (appserver_sub_id) REFERENCES appserver_sub(id) ON DELETE CASCADE,
    FOREIGN KEY (appserver_id) REFERENCES appserver(id) ON DELETE CASCADE,
    FOREIGN KEY (appuser_id) REFERENCES appuser(id) ON DELETE CASCADE,

    CONSTRAINT appserver_role_sub_uk_role_sub_server UNIQUE (appserver_role_id, appserver_sub_id, appserver_id),

    -- enforce that the sub really belongs to that server + user
    CONSTRAINT appserver_role_sub_uk_server_and_sub FOREIGN KEY (appserver_id, appserver_sub_id)
    REFERENCES appserver_sub(appserver_id, id) ON DELETE CASCADE,

  -- enforce that the role really belongs to that same server
    CONSTRAINT appserver_role_sub_uk_server_and_role FOREIGN KEY (appserver_id, appserver_role_id)
    REFERENCES appserver_role(appserver_id, id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS appserver_role_sub;
-- +goose StatementEnd
