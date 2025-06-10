package service

import (
	"mist/src/producer"
	"mist/src/psql_db/db"
)

type ServiceDeps struct {
	Db        db.Querier
	MProducer *producer.MProducer
}
