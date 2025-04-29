-- name: CreateChannelPermission :one
INSERT INTO channel_permission (
  channel_id,
  appserver_role_id,
  read_all,
  write_all
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetChannelPermissionById :one
SELECT *
FROM channel_permission
WHERE id = $1
LIMIT 1;

-- name: ListChannelPermissions :many
SELECT *
FROM channel_permission
WHERE channel_id = $1;

-- name: DeleteChannelPermission :execrows
DELETE FROM channel_permission
WHERE id = $1;
