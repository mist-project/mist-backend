package rpcs

import (
	"github.com/jackc/pgx/v5/pgxpool"

	pb_channel "mist/src/protos/channel/v1"
	pb_server "mist/src/protos/server/v1"
)

type AppserverGRPCService struct {
	pb_server.UnimplementedServerServiceServer
	DbcPool *pgxpool.Pool
}

type ChannelGRPCService struct {
	pb_channel.UnimplementedChannelServiceServer
	DbcPool *pgxpool.Pool
}
