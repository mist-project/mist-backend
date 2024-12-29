package rpcs

import (
	"github.com/jackc/pgx/v5/pgxpool"

	pb_servers "mist/src/protos/server/v1"
)

type Grpcserver struct {
	pb_servers.UnimplementedServerServiceServer
	DbcPool *pgxpool.Pool
}
