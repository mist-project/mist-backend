
-- name: GetAppserverRoleById :one
SELECT *
FROM appserver_role
WHERE id=$1
LIMIT 1;

-- name: CreateAppserverRole :one
INSERT INTO appserver_role (
  appserver_id,
  name,
  appserver_permission_mask,
  channel_permission_mask,
  sub_permission_mask
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5
)
RETURNING *;

-- name: ListAppserverRoles :many
SELECT *
FROM appserver_role
WHERE appserver_id=$1;

-- name: GetAppuserRoles :many
SELECT
  ar.id,
  ar.name,
  ar.appserver_permission_mask,
  ar.channel_permission_mask,
  ar.sub_permission_mask
FROM appserver_role AS ar
JOIN appserver_role_sub AS ars ON ars.appserver_role_id = ar.id
WHERE ars.appuser_id = $1
  AND ar.appserver_id = $2;


-- name: GetAppusersWithOnlySpecifiedRole :many
SELECT appuser.*
FROM appuser
JOIN appserver_role_sub ON appserver_role_sub.appuser_id = appuser.id
WHERE appserver_role_sub.appserver_role_id = $1
GROUP BY appuser.id
HAVING COUNT(*) = 1
   AND COUNT(*) FILTER (WHERE appserver_role_sub.appserver_role_id != $1) = 0;


-- name: DeleteAppserverRole :execrows
DELETE FROM appserver_role as ar
WHERE ar.id=$1;
