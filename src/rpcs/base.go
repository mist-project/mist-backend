package rpcs

import (
	"github.com/jackc/pgx/v5/pgxpool"

	pb_appserver "mist/src/protos/v1/appserver"
	pb_appuser "mist/src/protos/v1/appuser"
	pb_channel "mist/src/protos/v1/channel"
)

type AppserverGRPCService struct {
	pb_appserver.UnimplementedAppserverServiceServer
	DbcPool *pgxpool.Pool
}

type ChannelGRPCService struct {
	pb_channel.UnimplementedChannelServiceServer
	DbcPool *pgxpool.Pool
}

type AppuserGRPCService struct {
	pb_appuser.UnimplementedAppuserServiceServer
	DbcPool *pgxpool.Pool
}
