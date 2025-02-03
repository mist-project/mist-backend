// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: appserver_role_sub.sql

package qx

import (
	"context"

	"github.com/google/uuid"
)

const createAppserverRoleSub = `-- name: CreateAppserverRoleSub :one
INSERT INTO appserver_role_sub (
  appserver_sub_id,
  appserver_role_id,
  appuser_id
) VALUES (
  $1,
  $2,
  $3
)
RETURNING id, appuser_id, appserver_role_id, appserver_sub_id, created_at, updated_at
`

type CreateAppserverRoleSubParams struct {
	AppserverSubID  uuid.UUID
	AppserverRoleID uuid.UUID
	AppuserID       uuid.UUID
}

// --- APP SERVER ROLE SUBS -----
func (q *Queries) CreateAppserverRoleSub(ctx context.Context, arg CreateAppserverRoleSubParams) (AppserverRoleSub, error) {
	row := q.db.QueryRow(ctx, createAppserverRoleSub, arg.AppserverSubID, arg.AppserverRoleID, arg.AppuserID)
	var i AppserverRoleSub
	err := row.Scan(
		&i.ID,
		&i.AppuserID,
		&i.AppserverRoleID,
		&i.AppserverSubID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteAppserverRoleSub = `-- name: DeleteAppserverRoleSub :execrows
DELETE FROM appserver_role_sub as ars
USING appserver as a, appserver_role as ar
WHERE a.id=ar.appserver_id
  AND ar.id=ars.appserver_role_id
  AND ars.id=$1
  AND a.appuser_id=$2
`

type DeleteAppserverRoleSubParams struct {
	ID        uuid.UUID
	AppuserID uuid.UUID
}

func (q *Queries) DeleteAppserverRoleSub(ctx context.Context, arg DeleteAppserverRoleSubParams) (int64, error) {
	result, err := q.db.Exec(ctx, deleteAppserverRoleSub, arg.ID, arg.AppuserID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
