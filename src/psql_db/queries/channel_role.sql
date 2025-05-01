
-- name: GetChannelRoleById :one
SELECT *
FROM channel_role
WHERE id=$1
LIMIT 1;

-- name: CreateChannelRole :one
INSERT INTO channel_role (
  channel_id,
  appserver_role_id,
  appserver_id
) VALUES (
  $1,
  $2,
  $3
)
RETURNING *;

-- name: ListChannelRoles :many
SELECT *
FROM channel_role
WHERE channel_id=$1;

-- name: DeleteChannelRole :execrows
DELETE FROM channel_role as cr
WHERE cr.id=$1;
