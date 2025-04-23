package rpcs

import (
	"log"
	"mist/src/middleware"
	pb_appserver "mist/src/protos/v1/appserver"
	pb_appserverrole "mist/src/protos/v1/appserver_role"
	pb_appserverrolesub "mist/src/protos/v1/appserver_role_sub"
	pb_appserversub "mist/src/protos/v1/appserver_sub"
	pb_appuser "mist/src/protos/v1/appuser"
	pb_channel "mist/src/protos/v1/channel"
	"mist/src/psql_db/qx"

	"github.com/bufbuild/protovalidate-go"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"google.golang.org/grpc"
)

type AppserverGRPCService struct {
	pb_appserver.UnimplementedAppserverServiceServer
	DbConn qx.DBTX
}

type AppserverSubGRPCService struct {
	pb_appserversub.UnimplementedAppserverSubServiceServer
	DbConn qx.DBTX
}

type AppserverRoleGRPCService struct {
	pb_appserverrole.UnimplementedAppserverRoleServiceServer
	DbConn qx.DBTX
}

type AppserverRoleSubGRPCService struct {
	pb_appserverrolesub.UnimplementedAppserverRoleSubServiceServer
	DbConn qx.DBTX
}

type ChannelGRPCService struct {
	pb_channel.UnimplementedChannelServiceServer
	DbConn qx.DBTX
}

type AppuserGRPCService struct {
	pb_appuser.UnimplementedAppuserServiceServer
	Db qx.Querier
}

func RegisterGrpcServices(s *grpc.Server, dbConn qx.DBTX) {
	pb_appuser.RegisterAppuserServiceServer(s, &AppuserGRPCService{Db: qx.New(dbConn)})
	pb_appserver.RegisterAppserverServiceServer(s, &AppserverGRPCService{DbConn: dbConn})
	pb_appserversub.RegisterAppserverSubServiceServer(s, &AppserverSubGRPCService{DbConn: dbConn})
	pb_appserverrole.RegisterAppserverRoleServiceServer(s, &AppserverRoleGRPCService{DbConn: dbConn})
	pb_appserverrolesub.RegisterAppserverRoleSubServiceServer(s, &AppserverRoleSubGRPCService{DbConn: dbConn})
	pb_channel.RegisterChannelServiceServer(s, &ChannelGRPCService{DbConn: dbConn})
}

func BaseInterceptors() grpc.ServerOption {
	validator, err := protovalidate.New()
	if err != nil {
		log.Fatalf("failed to create protovalidate validator")
	}

	return grpc.ChainUnaryInterceptor(
		middleware.AuthJwtInterceptor,
		protovalidate_middleware.UnaryServerInterceptor(validator),
	)
}
