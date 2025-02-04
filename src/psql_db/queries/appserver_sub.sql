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
JOIN appserver as aps ON apssub.appserver_id=aps.id
WHERE apssub.appuser_id=$1;

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
