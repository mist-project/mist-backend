
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

-- name: ListAppserverRoles :many
SELECT *
FROM appserver_role
WHERE appserver_id=$1;

-- name: DeleteAppserverRole :execrows
DELETE FROM appserver_role as ar
WHERE ar.id=$1;
