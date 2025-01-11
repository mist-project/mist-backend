-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS appserver_sub (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    appserver_id UUID NOT NULL,
    app_user_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    FOREIGN KEY (appserver_id) REFERENCES appserver(id) ON DELETE CASCADE,
    CONSTRAINT appserver_sub_uk_appserver_owner UNIQUE (appserver_id, app_user_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS appserver_sub;
-- +goose StatementEnd
