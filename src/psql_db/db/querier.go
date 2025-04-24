package db

import (
	"mist/src/psql_db/qx"

	"github.com/jackc/pgx/v5"
)

type Querier interface {
	qx.Querier
	WithTx(tx pgx.Tx) Querier
}

type Queries struct {
	*qx.Queries
}

func (q *Queries) WithTx(tx pgx.Tx) Querier {
	return &Queries{
		Queries: q.Queries.WithTx(tx),
	}
}

func NewQuerier(q *qx.Queries) Querier {
	return &Queries{
		Queries: q,
	}
}
