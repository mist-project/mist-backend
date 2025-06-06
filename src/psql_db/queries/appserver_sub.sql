-- name: GetAppserverSubById :one
SELECT *
FROM appserver_sub
WHERE id=$1
LIMIT 1;

-- name: CreateAppserverSub :one
INSERT INTO appserver_sub (
  appserver_id,
  appuser_id
) VALUES (
  $1,
  $2
)
RETURNING *;

-- name: ListUserServerSubs :many
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

-- name: ListAppserverUserSubs :many
SELECT
  asub.id as appserver_sub_id,
  auser.id as appuser_id,
  auser.username as appuser_username,
  auser.created_at as appuser_created_at,
  auser.updated_at as appuser_updated_at
FROM appserver_sub as asub
JOIN appuser as auser ON asub.appuser_id=auser.id
WHERE asub.appserver_id=$1;

-- name: FilterAppserverSub :many
SELECT 
  sub.id,
  sub.appuser_id,
  sub.appserver_id,
  sub.created_at,
  sub.updated_at
FROM appserver_sub as sub
WHERE appuser_id=COALESCE(sqlc.narg('appuser_id'), appuser_id)
  AND appserver_id=COALESCE(sqlc.narg('appserver_id'), appserver_id);


-- name: DeleteAppserverSub :execrows
DELETE FROM appserver_sub
WHERE id=$1;
