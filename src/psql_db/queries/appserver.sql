
-- name: GetAppserverById :one
SELECT *
FROM appserver
WHERE id=$1
LIMIT 1;

-- name: CreateAppserver :one
INSERT INTO appserver (
  name,
  appuser_id
) VALUES (
  $1,
  $2
)
RETURNING *;

-- name: ListUserAppservers :many
SELECT *
FROM appserver
WHERE name=COALESCE(sqlc.narg('name'), name)
  AND appuser_id = $1; -- This query might be removed. Hence the 1=0. So it returns no data.



-- name: DeleteAppserver :execrows
DELETE FROM appserver
WHERE id=$1
  AND appuser_id=$2;