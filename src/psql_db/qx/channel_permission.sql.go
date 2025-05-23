// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: channel_permission.sql

package qx

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createChannelPermission = `-- name: CreateChannelPermission :one
INSERT INTO channel_permission (
  channel_id,
  appserver_role_id,
  read_all,
  write_all
) VALUES (
  $1, $2, $3, $4
)
RETURNING id, channel_id, appserver_role_id, read_all, write_all, created_at, updated_at
`

type CreateChannelPermissionParams struct {
	ChannelID       uuid.UUID
	AppserverRoleID uuid.UUID
	ReadAll         pgtype.Bool
	WriteAll        pgtype.Bool
}

func (q *Queries) CreateChannelPermission(ctx context.Context, arg CreateChannelPermissionParams) (ChannelPermission, error) {
	row := q.db.QueryRow(ctx, createChannelPermission,
		arg.ChannelID,
		arg.AppserverRoleID,
		arg.ReadAll,
		arg.WriteAll,
	)
	var i ChannelPermission
	err := row.Scan(
		&i.ID,
		&i.ChannelID,
		&i.AppserverRoleID,
		&i.ReadAll,
		&i.WriteAll,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteChannelPermission = `-- name: DeleteChannelPermission :execrows
DELETE FROM channel_permission
WHERE id = $1
`

func (q *Queries) DeleteChannelPermission(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.Exec(ctx, deleteChannelPermission, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const getChannelPermissionById = `-- name: GetChannelPermissionById :one
SELECT id, channel_id, appserver_role_id, read_all, write_all, created_at, updated_at
FROM channel_permission
WHERE id = $1
LIMIT 1
`

func (q *Queries) GetChannelPermissionById(ctx context.Context, id uuid.UUID) (ChannelPermission, error) {
	row := q.db.QueryRow(ctx, getChannelPermissionById, id)
	var i ChannelPermission
	err := row.Scan(
		&i.ID,
		&i.ChannelID,
		&i.AppserverRoleID,
		&i.ReadAll,
		&i.WriteAll,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listChannelPermissions = `-- name: ListChannelPermissions :many
SELECT id, channel_id, appserver_role_id, read_all, write_all, created_at, updated_at
FROM channel_permission
WHERE channel_id = $1
`

func (q *Queries) ListChannelPermissions(ctx context.Context, channelID uuid.UUID) ([]ChannelPermission, error) {
	rows, err := q.db.Query(ctx, listChannelPermissions, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ChannelPermission
	for rows.Next() {
		var i ChannelPermission
		if err := rows.Scan(
			&i.ID,
			&i.ChannelID,
			&i.AppserverRoleID,
			&i.ReadAll,
			&i.WriteAll,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
