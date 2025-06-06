-- name: GetAppserverRoleSubById :one
SELECT *
FROM appserver_role_sub
WHERE id=$1
LIMIT 1;

-- name: CreateAppserverRoleSub :one
INSERT INTO appserver_role_sub (
  appserver_sub_id,
  appserver_role_id,
  appuser_id,
  appserver_id
) VALUES (
  $1,
  $2,
  $3,
  $4
)
RETURNING *;

-- name: ListServerRoleSubs :many
SELECT
  role_sub.id,
  role_sub.appuser_id,
  role_sub.appserver_role_id,
  role_sub.appserver_id

FROM appserver_role_sub AS role_sub
WHERE role_sub.appserver_id=$1;

-- name: FilterAppserverRoleSub :many
SELECT
  role_sub.id,
  role_sub.appuser_id,
  role_sub.appserver_role_id,
  role_sub.appserver_id
FROM appserver_role_sub AS role_sub
WHERE appuser_id=COALESCE(sqlc.narg('appuser_id'), appuser_id)
  AND appserver_id=COALESCE(sqlc.narg('appserver_id'), appserver_id)
  AND appserver_role_id=COALESCE(sqlc.narg('appserver_role_id'), appserver_role_id)
  AND appserver_sub_id=COALESCE(sqlc.narg('appserver_sub_id'), appserver_sub_id);

-- name: DeleteAppserverRoleSub :execrows
DELETE FROM appserver_role_sub as ars
WHERE ars.id=$1;
