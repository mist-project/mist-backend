----- APP USER QUERIES -----
-- name: GetAppUser :one
SELECT *
FROM app_user
WHERE id=$1
LIMIT 1;

-- name: CreateAppUser :one
INSERT INTO app_user (
  id, username
) VALUES ($1, $2)
RETURNING *;

----- APP SERVER QUERIES -----
-- name: GetAppserver :one
SELECT *
FROM appserver
WHERE id=$1
LIMIT 1;

-- name: ListUserAppservers :many
SELECT *
FROM appserver
WHERE name=COALESCE(sqlc.narg('name'), name)
  AND app_user_id = $1; -- This query might be removed. Hence the 1=0. So it returns no data.

-- name: CreateAppserver :one
INSERT INTO appserver (
  name,
  app_user_id
) VALUES (
  $1,
  $2
)
RETURNING *;

-- name: DeleteAppserver :execrows
DELETE FROM appserver
WHERE id=$1
  AND app_user_id=$2;

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
WHERE apssub.app_user_id=$1;

-- name: CreateAppserverSub :one
INSERT INTO appserver_sub (
  appserver_id,
  app_user_id
) VALUES (
  $1,
  $2
)
RETURNING *;

-- name: DeleteAppserverSub :execrows
DELETE FROM appserver_sub
WHERE id=$1;

----- APP SERVER ROLES -----
-- name: GetAppserverRoles :many
SELECT *
FROM appserver_role
WHERE appserver_id=$1;

-- name: CreateAppserverRole :one
INSERT INTO appserver_role (
  appserver_id,
  name
) VALUES (
  $1,
  $2
)
RETURNING *;

-- name: DeleteAppserverRole :execrows
DELETE FROM appserver_role as ar
USING appserver as a 
WHERE a.id=ar.appserver_id
  AND ar.id=$1
  AND a.app_user_id=$2;

----- APP SERVER ROLE SUBS -----
-- name: CreateAppserverRoleSub :one
INSERT INTO appserver_role_sub (
  appserver_sub_id,
  appserver_role_id,
  app_user_id
) VALUES (
  $1,
  $2,
  $3
)
RETURNING *;

-- name: DeleteAppserverRoleSub :execrows
DELETE FROM appserver_role_sub as ars
USING appserver as a, appserver_role as ar
WHERE a.id=ar.appserver_id
  AND ar.id=ars.appserver_role_id
  AND ars.id=$1
  AND a.app_user_id=$2;

----- CHANNEL QUERIES -----
-- name: GetChannel :one
SELECT *
FROM channel
WHERE id=$1
LIMIT 1;

-- name: ListChannels :many
SELECT *
FROM channel
WHERE name=COALESCE(sqlc.narg('name'), name)
  AND appserver_id=COALESCE(sqlc.narg('appserver_id'), appserver_id);

-- name: CreateChannel :one
INSERT INTO channel (
  name,
  appserver_id
) VALUES (
  $1,
  $2
)
RETURNING *;

-- name: DeleteChannel :execrows
DELETE FROM channel
WHERE id=$1;
