-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS app_user (
    id UUID PRIMARY KEY NOT NULL,
    username VARCHAR(255) NOT NULL,
    online BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

ALTER TABLE appserver
ADD CONSTRAINT fk_app_user
FOREIGN KEY (app_user_id)
REFERENCES app_user(id)
ON DELETE CASCADE;

ALTER TABLE appserver_sub
ADD CONSTRAINT fk_app_user
FOREIGN KEY (app_user_id)
REFERENCES app_user(id)
ON DELETE CASCADE;

ALTER TABLE appserver_role_sub
ADD CONSTRAINT fk_app_user
FOREIGN KEY (app_user_id)
REFERENCES app_user(id)
ON DELETE CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE appserver_role_sub
DROP CONSTRAINT IF EXISTS fk_app_user;

ALTER TABLE appserver_sub
DROP CONSTRAINT IF EXISTS fk_app_user;

ALTER TABLE appserver
DROP CONSTRAINT IF EXISTS fk_app_user;

DROP TABLE IF EXISTS app_user;
-- +goose StatementEnd
