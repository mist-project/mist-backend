package rpcs

import (
	pb_appserver "mist/src/protos/v1/appserver"
	pb_appuser "mist/src/protos/v1/appuser"
	pb_channel "mist/src/protos/v1/channel"
	"mist/src/psql_db/qx"
)

type AppserverGRPCService struct {
	pb_appserver.UnimplementedAppserverServiceServer
	DbConn qx.DBTX
}

type ChannelGRPCService struct {
	pb_channel.UnimplementedChannelServiceServer
	DbConn qx.DBTX
}

type AppuserGRPCService struct {
	pb_appuser.UnimplementedAppuserServiceServer
	DbConn qx.DBTX
}
