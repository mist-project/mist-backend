
-- name: GetAppserverRoleById :one
SELECT *
FROM appserver_role
WHERE id=$1
LIMIT 1;

-- name: CreateAppserverRole :one
INSERT INTO appserver_role (
  appserver_id,
  name
) VALUES (
  $1,
  $2
)
RETURNING *;

-- name: GetAppserverRoles :many
SELECT *
FROM appserver_role
WHERE appserver_id=$1;

-- name: DeleteAppserverRole :execrows
DELETE FROM appserver_role as ar
USING appserver as a 
WHERE a.id=ar.appserver_id
  AND ar.id=$1
  AND a.appuser_id=$2;
