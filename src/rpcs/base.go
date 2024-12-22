package rpcs

import (
	"github.com/jackc/pgx/v5/pgxpool"

	pb_mistbe "mist/src/protos/mistbe/v1"
)

type Grpcserver struct {
	pb_mistbe.UnimplementedMistBEServiceServer
	DbcPool *pgxpool.Pool
}
