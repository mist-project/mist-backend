package rpcs

import (
	"github.com/jackc/pgx/v5/pgxpool"

	pb_channel "mist/src/protos/v1/channel"
	pb_server "mist/src/protos/v1/server"
)

type AppserverGRPCService struct {
	pb_server.UnimplementedServerServiceServer
	DbcPool *pgxpool.Pool
}

type ChannelGRPCService struct {
	pb_channel.UnimplementedChannelServiceServer
	DbcPool *pgxpool.Pool
}
