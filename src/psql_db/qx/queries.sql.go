// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: queries.sql

package qx

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createAppserver = `-- name: CreateAppserver :one

INSERT INTO appserver (
  name,
  owner_id
) VALUES (
  $1,
  $2
)
RETURNING id, name, owner_id, created_at, updated_at
`

type CreateAppserverParams struct {
	Name    string
	OwnerID uuid.UUID
}

// This query might be removed. Hence the 1=0. So it returns no data.
func (q *Queries) CreateAppserver(ctx context.Context, arg CreateAppserverParams) (Appserver, error) {
	row := q.db.QueryRow(ctx, createAppserver, arg.Name, arg.OwnerID)
	var i Appserver
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.OwnerID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createAppserverRole = `-- name: CreateAppserverRole :one
INSERT INTO appserver_role (
  appserver_id,
  name
) VALUES (
  $1,
  $2
)
RETURNING id, appserver_id, name, created_at, updated_at
`

type CreateAppserverRoleParams struct {
	AppserverID uuid.UUID
	Name        string
}

func (q *Queries) CreateAppserverRole(ctx context.Context, arg CreateAppserverRoleParams) (AppserverRole, error) {
	row := q.db.QueryRow(ctx, createAppserverRole, arg.AppserverID, arg.Name)
	var i AppserverRole
	err := row.Scan(
		&i.ID,
		&i.AppserverID,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createAppserverSub = `-- name: CreateAppserverSub :one
INSERT INTO appserver_sub (
  appserver_id,
  owner_id
) VALUES (
  $1,
  $2
)
RETURNING id, appserver_id, owner_id, created_at, updated_at
`

type CreateAppserverSubParams struct {
	AppserverID uuid.UUID
	OwnerID     uuid.UUID
}

func (q *Queries) CreateAppserverSub(ctx context.Context, arg CreateAppserverSubParams) (AppserverSub, error) {
	row := q.db.QueryRow(ctx, createAppserverSub, arg.AppserverID, arg.OwnerID)
	var i AppserverSub
	err := row.Scan(
		&i.ID,
		&i.AppserverID,
		&i.OwnerID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createChannel = `-- name: CreateChannel :one
INSERT INTO channel (
  name,
  appserver_id
) VALUES (
  $1,
  $2
)
RETURNING id, name, appserver_id, created_at, updated_at
`

type CreateChannelParams struct {
	Name        string
	AppserverID uuid.UUID
}

func (q *Queries) CreateChannel(ctx context.Context, arg CreateChannelParams) (Channel, error) {
	row := q.db.QueryRow(ctx, createChannel, arg.Name, arg.AppserverID)
	var i Channel
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.AppserverID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteAppserver = `-- name: DeleteAppserver :execrows
DELETE
FROM appserver
WHERE id=$1 AND owner_id=$2
`

type DeleteAppserverParams struct {
	ID      uuid.UUID
	OwnerID uuid.UUID
}

func (q *Queries) DeleteAppserver(ctx context.Context, arg DeleteAppserverParams) (int64, error) {
	result, err := q.db.Exec(ctx, deleteAppserver, arg.ID, arg.OwnerID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const deleteAppserverRole = `-- name: DeleteAppserverRole :execrows
DELETE
FROM appserver_role as ar
USING appserver as a 
WHERE a.id=ar.appserver_id AND ar.id=$1 AND a.owner_id=$2
`

type DeleteAppserverRoleParams struct {
	ID      uuid.UUID
	OwnerID uuid.UUID
}

func (q *Queries) DeleteAppserverRole(ctx context.Context, arg DeleteAppserverRoleParams) (int64, error) {
	result, err := q.db.Exec(ctx, deleteAppserverRole, arg.ID, arg.OwnerID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const deleteAppserverSub = `-- name: DeleteAppserverSub :execrows
DELETE
FROM appserver_sub
WHERE id=$1
`

func (q *Queries) DeleteAppserverSub(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.Exec(ctx, deleteAppserverSub, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const deleteChannel = `-- name: DeleteChannel :execrows
DELETE
FROM channel
WHERE id=$1
`

func (q *Queries) DeleteChannel(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.Exec(ctx, deleteChannel, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const getAppserver = `-- name: GetAppserver :one
SELECT id, name, owner_id, created_at, updated_at
FROM appserver
WHERE id=$1
LIMIT 1
`

// --- APP SERVER QUERIES -----
func (q *Queries) GetAppserver(ctx context.Context, id uuid.UUID) (Appserver, error) {
	row := q.db.QueryRow(ctx, getAppserver, id)
	var i Appserver
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.OwnerID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getAppserverRoles = `-- name: GetAppserverRoles :many
SELECT id, appserver_id, name, created_at, updated_at
FROM appserver_role
WHERE appserver_id=$1
`

// --- APP SERVER ROLES -----
func (q *Queries) GetAppserverRoles(ctx context.Context, appserverID uuid.UUID) ([]AppserverRole, error) {
	rows, err := q.db.Query(ctx, getAppserverRoles, appserverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AppserverRole
	for rows.Next() {
		var i AppserverRole
		if err := rows.Scan(
			&i.ID,
			&i.AppserverID,
			&i.Name,
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

const getAppserverSub = `-- name: GetAppserverSub :one
SELECT id, appserver_id, owner_id, created_at, updated_at
FROM appserver_sub
WHERE id=$1
LIMIT 1
`

// --- APPSERVER SUB QUERIES -----
func (q *Queries) GetAppserverSub(ctx context.Context, id uuid.UUID) (AppserverSub, error) {
	row := q.db.QueryRow(ctx, getAppserverSub, id)
	var i AppserverSub
	err := row.Scan(
		&i.ID,
		&i.AppserverID,
		&i.OwnerID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getChannel = `-- name: GetChannel :one
SELECT id, name, appserver_id, created_at, updated_at
FROM channel
WHERE id=$1
LIMIT 1
`

// --- CHANNEL QUERIES -----
func (q *Queries) GetChannel(ctx context.Context, id uuid.UUID) (Channel, error) {
	row := q.db.QueryRow(ctx, getChannel, id)
	var i Channel
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.AppserverID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserAppserverSubs = `-- name: GetUserAppserverSubs :many
SELECT 
  apssub.id as appserver_sub_id,
  aps.id,
  aps.name,
  aps.created_at,
  aps.updated_at  
FROM appserver_sub as apssub
JOIN appserver as aps ON apssub.appserver_id=aps.id
WHERE
  apssub.owner_id=$1
`

type GetUserAppserverSubsRow struct {
	AppserverSubID uuid.UUID
	ID             uuid.UUID
	Name           string
	CreatedAt      pgtype.Timestamp
	UpdatedAt      pgtype.Timestamp
}

func (q *Queries) GetUserAppserverSubs(ctx context.Context, ownerID uuid.UUID) ([]GetUserAppserverSubsRow, error) {
	rows, err := q.db.Query(ctx, getUserAppserverSubs, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUserAppserverSubsRow
	for rows.Next() {
		var i GetUserAppserverSubsRow
		if err := rows.Scan(
			&i.AppserverSubID,
			&i.ID,
			&i.Name,
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

const listAppservers = `-- name: ListAppservers :many
SELECT id, name, owner_id, created_at, updated_at
FROM appserver
WHERE
  name=COALESCE($1, name)
  AND
  1=0
`

func (q *Queries) ListAppservers(ctx context.Context, name pgtype.Text) ([]Appserver, error) {
	rows, err := q.db.Query(ctx, listAppservers, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Appserver
	for rows.Next() {
		var i Appserver
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.OwnerID,
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

const listChannels = `-- name: ListChannels :many
SELECT id, name, appserver_id, created_at, updated_at
FROM channel
WHERE
  (name=COALESCE($1, name))
  AND
  (appserver_id=COALESCE($2, appserver_id))
`

type ListChannelsParams struct {
	Name        pgtype.Text
	AppserverID pgtype.UUID
}

func (q *Queries) ListChannels(ctx context.Context, arg ListChannelsParams) ([]Channel, error) {
	rows, err := q.db.Query(ctx, listChannels, arg.Name, arg.AppserverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Channel
	for rows.Next() {
		var i Channel
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.AppserverID,
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
