-- name: CreateChannel :one
INSERT INTO channel (
  name,
  appserver_id,
  is_private
) VALUES (
  $1,
  $2,
  $3
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
WHERE channel_role.appserver_role_id = ANY($1::uuid[]);

-- name: GetChannelsForUser :many
SELECT DISTINCT channel.*
FROM channel
LEFT JOIN channel_role ON channel_role.channel_id = channel.id
LEFT JOIN appserver_role_sub ON appserver_role_sub.app_server_role_id = channel_role.appserver_role_id
WHERE channel.appserver_id = $2
  AND (
    channel_role.id IS NULL -- channels with no roles
    OR appserver_role_sub.appuser_id = $1 -- channels where user has a role
  );

-- name: FilterChannel :many
SELECT *
FROM channel
WHERE appserver_id = COALESCE(sqlc.narg('appserver_id'), appserver_id)
  AND is_private = COALESCE(sqlc.narg('is_private'), is_private);


-- name: DeleteChannel :execrows
DELETE FROM channel
WHERE id=$1;
