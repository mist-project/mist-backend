-- name: GetAppserver :one
SELECT *
FROM appserver
WHERE id=$1
LIMIT 1;

-- name: ListAppservers :many
SELECT *
FROM appserver
WHERE
  (name = sqlc.narg('name') OR sqlc.narg('name') IS NULL);
;

-- name: CreateAppserver :one
INSERT INTO appserver (
  name
) values (
  $1
)
RETURNING *;

-- name: DeleteAppserver :execrows
DELETE FROM appserver
WHERE id = $1;