-- name: CreateChannel :one
INSERT INTO channel (
  name,
  appserver_id
) VALUES (
  $1,
  $2
)
RETURNING *;

-- name: GetChannelById :one
SELECT *
FROM channel
WHERE id=$1
LIMIT 1;

-- name: ListServerChannels :many
SELECT *
FROM channel
WHERE name=COALESCE(sqlc.narg('name'), name)
  AND appserver_id=$1;


-- name: GetChannelUsersByRoles :many
SELECT DISTINCT appuser.*
FROM appuser
JOIN appserver_role_sub ON appserver_role_sub.appuser_id = appuser.id
JOIN channel_role ON channel_role.appserver_role_id = appserver_role_sub.app_server_role_id
WHERE channel_role.id = ANY($1::uuid[]);

-- name: DeleteChannel :execrows
DELETE FROM channel
WHERE id=$1;
