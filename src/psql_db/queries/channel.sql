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

-- name: GetChannelsIdIn :many
SELECT *
FROM channel
WHERE id = ANY($1::uuid[]);

-- name: ListServerChannels :many
SELECT *
FROM channel
WHERE name=COALESCE(sqlc.narg('name'), name)
  AND appserver_id=$1;


-- name: GetChannelsForUsers :many
SELECT DISTINCT
  u.appuser_id::uuid as appuser_id,
  channel.id AS channel_id,
  channel.name AS channel_name,
  channel.is_private AS channel_is_private,
  channel.appserver_id AS channel_appserver_id
FROM (
  SELECT unnest($1::uuid[]) AS appuser_id
) u
LEFT JOIN channel
  ON channel.appserver_id = $2
LEFT JOIN channel_role
  ON channel_role.channel_id = channel.id
LEFT JOIN appserver_role_sub
  ON appserver_role_sub.appserver_role_id = channel_role.appserver_role_id
    AND appserver_role_sub.appuser_id = u.appuser_id
WHERE
  channel.is_private = false
  OR appserver_role_sub.appuser_id IS NOT NULL
GROUP BY (u.appuser_id, channel.id);


-- name: FilterChannel :many
SELECT *
FROM channel
WHERE appserver_id = COALESCE(sqlc.narg('appserver_id'), appserver_id)
  AND is_private = COALESCE(sqlc.narg('is_private'), is_private);


-- name: DeleteChannel :execrows
DELETE FROM channel
WHERE id=$1;
