package db

import (
	"context"
	"fmt"
	"log/slog"
	"mist/src/faults"
	"mist/src/psql_db/qx"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

type Querier interface {
	qx.Querier
	Begin(ctx context.Context) (Querier, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Queries struct {
	*qx.Queries
	dbConn DBTX
}

func (q *Queries) Begin(ctx context.Context) (Querier, error) {
	db, ok := q.dbConn.(*pgxpool.Pool)

	if !ok {
		// If dbConn is already a transaction, return the same querier
		if _, ok := q.dbConn.(pgx.Tx); ok {
			return q, nil
		}

		return nil, faults.DatabaseError(
			"querier's dbConn is not a pgxpool.Pool or pgx.Tx, cannot begin transaction", slog.LevelError,
		)
	}

	// Begin a new transaction
	tx, err := db.Begin(ctx)

	if err != nil {
		return nil, faults.DatabaseError(fmt.Sprintf("failed to begin transaction %v", err), slog.LevelError)
	}

	// Return a new Queries instance with the transaction
	return &Queries{
		Queries: q.Queries.WithTx(tx),
		dbConn:  tx,
	}, nil
}

func (q *Queries) Commit(ctx context.Context) error {
	tx, ok := q.dbConn.(pgx.Tx)
	if !ok {
		return faults.DatabaseError("querier's dbConn is not a transaction, cannot commit", slog.LevelError)
	}

	if err := tx.Commit(ctx); err != nil {
		return faults.DatabaseError(fmt.Sprintf("failed to commit transaction %v", err), slog.LevelError)
	}

	return nil
}

func (q *Queries) Rollback(ctx context.Context) error {
	tx, ok := q.dbConn.(pgx.Tx)
	if !ok {
		return faults.DatabaseError("querier's dbConn is not a transaction, cannot rollback", slog.LevelError)
	}

	if err := tx.Rollback(ctx); err != nil {
		return faults.DatabaseError(fmt.Sprintf("failed to rollback transaction %v", err), slog.LevelError)
	}

	return nil
}

func NewQuerier(dbConn DBTX) Querier {
	return &Queries{
		Queries: qx.New(dbConn),
		dbConn:  dbConn,
	}
}
