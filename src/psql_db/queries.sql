----- APP SERVER QUERIES -----
-- name: GetAppserver :one
SELECT *
FROM appserver
WHERE id=$1
LIMIT 1;

-- name: ListAppservers :many
SELECT *
FROM appserver
WHERE
  name = COALESCE(sqlc.narg('name'), name);

-- name: CreateAppserver :one
INSERT INTO appserver (
  name
) values (
  $1
)
RETURNING *;

-- name: DeleteAppserver :execrows
DELETE FROM appserver
WHERE id = $1;

----- CHANNEL QUERIES -----
-- name: GetChannel :one
SELECT *
FROM channel
WHERE id=$1
LIMIT 1;

-- name: ListChannels :many
SELECT *
FROM channel
WHERE
  (name = COALESCE(sqlc.narg('name'), name))
  AND
  (appserver_id = COALESCE(sqlc.narg('appserver_id'), appserver_id));

-- name: CreateChannel :one
INSERT INTO channel (
  name,
  appserver_id
) values (
  $1,
  $2
)
RETURNING *;

-- name: DeleteChannel :execrows
DELETE FROM channel
WHERE id = $1;
