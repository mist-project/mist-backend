----- APP SERVER ROLE SUBS -----
-- name: CreateAppserverRoleSub :one
INSERT INTO appserver_role_sub (
  appserver_sub_id,
  appserver_role_id,
  appuser_id
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
  AND a.appuser_id=$2;
