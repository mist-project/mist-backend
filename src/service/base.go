package service

import (
	"mist/src/producer"
	"mist/src/psql_db/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ServiceDeps struct {
	Db        db.Querier
	DbConn    *pgxpool.Pool
	MProducer *producer.MProducer
}
