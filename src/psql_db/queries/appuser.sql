-- ----- APP USER QUERIES -----
-- name: GetAppuser :one
SELECT *
FROM appuser
WHERE id=$1
LIMIT 1;

-- name: CreateAppuser :one
INSERT INTO appuser (
  id,
  username
) VALUES (
  $1,
  $2
)
RETURNING *;
