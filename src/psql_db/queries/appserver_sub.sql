----- APPSERVER SUB QUERIES -----
-- name: GetAppserverSub :one
SELECT *
FROM appserver_sub
WHERE id=$1
LIMIT 1;

-- name: GetUserAppserverSubs :many
SELECT
  asub.id as appserver_sub_id,
  asub.appuser_id,
  aserver.id,
  aserver.name,
  aserver.created_at,
  aserver.updated_at  
FROM appserver_sub as asub
JOIN appserver as aserver ON asub.appserver_id=aserver.id
WHERE asub.appuser_id=$1;

-- name: GetAllUsersAppserverSubs :many
SELECT
  asub.id as appserver_sub_id,
  auser.id,
  auser.username,
  auser.created_at,
  auser.updated_at  
FROM appserver_sub as asub
JOIN appuser as auser ON asub.appuser_id=auser.id
WHERE asub.appserver_id=$1;

-- name: CreateAppserverSub :one
INSERT INTO appserver_sub (
  appserver_id,
  appuser_id
) VALUES (
  $1,
  $2
)
RETURNING *;

-- name: DeleteAppserverSub :execrows
DELETE FROM appserver_sub
WHERE id=$1;
