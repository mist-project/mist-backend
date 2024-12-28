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

----- APPSERVER SUB QUERIES -----
-- name: GetAppserverSub :one
SELECT *
FROM appserver_sub
WHERE id=$1
LIMIT 1;

-- name: GetUserAppserverSubs :many
SELECT 
  apssub.id as appserver_sub_id,
  aps.id,
  aps.name,
  aps.created_at,
  aps.updated_at  
FROM appserver_sub as apssub
JOIN appserver as aps ON apssub.appserver_id = aps.id
WHERE
  apssub.owner_id = $1;

-- name: CreateAppserverSub :one
INSERT INTO appserver_sub (
  appserver_id,
  owner_id
) values (
  $1,
  $2
)
RETURNING *;

-- name: DeleteAppserverSub :execrows
DELETE FROM appserver_sub
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
