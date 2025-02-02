// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package qx

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type AppUser struct {
	ID        uuid.UUID
	Username  string
	Online    bool
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
}

type Appserver struct {
	ID        uuid.UUID
	Name      string
	AppUserID uuid.UUID
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
}

type AppserverRole struct {
	ID          uuid.UUID
	AppserverID uuid.UUID
	Name        string
	CreatedAt   pgtype.Timestamp
	UpdatedAt   pgtype.Timestamp
}

type AppserverRoleSub struct {
	ID              uuid.UUID
	AppUserID       uuid.UUID
	AppserverRoleID uuid.UUID
	AppserverSubID  uuid.UUID
	CreatedAt       pgtype.Timestamp
	UpdatedAt       pgtype.Timestamp
}

type AppserverSub struct {
	ID          uuid.UUID
	AppserverID uuid.UUID
	AppUserID   uuid.UUID
	CreatedAt   pgtype.Timestamp
	UpdatedAt   pgtype.Timestamp
}

type Channel struct {
	ID          uuid.UUID
	Name        string
	AppserverID uuid.UUID
	CreatedAt   pgtype.Timestamp
	UpdatedAt   pgtype.Timestamp
}

type GooseDbVersion struct {
	ID        int32
	VersionID int64
	IsApplied bool
	Tstamp    pgtype.Timestamp
}
