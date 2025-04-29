-- name: CreateAppserverPermission :one
INSERT INTO appserver_permission (
  appserver_id,
  appuser_id,
  read_all,
  write_all,
  delete_all
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetAppserverPermissionById :one
SELECT *
FROM appserver_permission
WHERE id = $1
LIMIT 1;

-- name: GetAppserverPermissionForUser :one
SELECT *
FROM appserver_permission
WHERE appserver_id = $1
  AND appuser_id = $2
LIMIT 1;

-- name: ListAppserverPermissions :many
SELECT *
FROM appserver_permission
WHERE appserver_id = $1;

-- name: DeleteAppserverPermission :execrows
DELETE FROM appserver_permission
WHERE id = $1;
