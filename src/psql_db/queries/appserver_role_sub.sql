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

-- name: DeleteAppserverRoleSub :execrows
DELETE FROM appserver_role_sub AS role_sub
USING appserver AS a, appserver_role AS ar
WHERE a.id=ar.appserver_id
  AND ar.id=role_sub.appserver_role_id
  AND role_sub.id=$1
  AND a.appuser_id=$2;
