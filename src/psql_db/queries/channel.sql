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
  AND appserver_id=COALESCE(sqlc.narg('appserver_id'), appserver_id);


-- name: DeleteChannel :execrows
DELETE FROM channel
WHERE id=$1;
