package rpcs

import (
	"github.com/bufbuild/protovalidate-go"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc"

	"mist/src/middleware"
	"mist/src/permission"
	"mist/src/producer"
	"mist/src/protos/v1/appserver"
	"mist/src/protos/v1/appserver_role"
	"mist/src/protos/v1/appserver_role_sub"
	"mist/src/protos/v1/appserver_sub"
	"mist/src/protos/v1/appuser"
	"mist/src/protos/v1/channel"
	"mist/src/protos/v1/channel_role"
	"mist/src/psql_db/db"
)

type GrpcDependencies struct {
	Db        db.Querier
	DbTx      pgx.Tx
	MProducer *producer.MProducer
}

type AppuserGRPCService struct {
	appuser.UnimplementedAppuserServiceServer
	Deps *GrpcDependencies
}

type AppserverGRPCService struct {
	appserver.UnimplementedAppserverServiceServer
	Auth permission.Authorizer
	Deps *GrpcDependencies
}

type AppserverSubGRPCService struct {
	appserver_sub.UnimplementedAppserverSubServiceServer
	Auth permission.Authorizer
	Deps *GrpcDependencies
}

type AppserverRoleGRPCService struct {
	appserver_role.UnimplementedAppserverRoleServiceServer
	Auth permission.Authorizer
	Deps *GrpcDependencies
}

type AppserverRoleSubGRPCService struct {
	appserver_role_sub.UnimplementedAppserverRoleSubServiceServer
	Auth permission.Authorizer
	Deps *GrpcDependencies
}

type ChannelGRPCService struct {
	channel.UnimplementedChannelServiceServer
	Auth permission.Authorizer
	Deps *GrpcDependencies
}

type ChannelRoleGRPCService struct {
	channel_role.UnimplementedChannelRoleServiceServer
	Auth permission.Authorizer
	Deps *GrpcDependencies
}

func RegisterGrpcServices(s *grpc.Server, deps *GrpcDependencies) {

	// ----- APPUSER -----
	appuser.RegisterAppuserServiceServer(
		s,
		&AppuserGRPCService{
			Deps: deps,
		},
	)

	// ----- APPSERVER -----
	appserver.RegisterAppserverServiceServer(
		s,
		&AppserverGRPCService{
			Deps: deps,
			Auth: permission.NewAppserverAuthorizer(deps.Db),
		},
	)

	// ----- APPSERVER ROLE -----
	appserver_role.RegisterAppserverRoleServiceServer(
		s,
		&AppserverRoleGRPCService{
			Deps: deps,
			Auth: permission.NewAppserverRoleAuthorizer(deps.Db),
		},
	)

	// ----- APPSERVER ROLE SUB -----
	appserver_role_sub.RegisterAppserverRoleSubServiceServer(
		s,
		&AppserverRoleSubGRPCService{
			Deps: deps,
			Auth: permission.NewAppserverRoleSubAuthorizer(deps.Db),
		},
	)

	// ----- APPSERVER SUB -----
	appserver_sub.RegisterAppserverSubServiceServer(
		s,
		&AppserverSubGRPCService{
			Deps: deps,
			Auth: permission.NewAppserverSubAuthorizer(deps.Db),
		},
	)

	// ----- CHANNEL -----
	channel.RegisterChannelServiceServer(
		s,
		&ChannelGRPCService{
			Deps: deps,
			Auth: permission.NewChannelAuthorizer(deps.Db),
		},
	)

	// ----- CHANNEL ROLE -----
	channel_role.RegisterChannelRoleServiceServer(
		s,
		&ChannelRoleGRPCService{
			Deps: deps,
			Auth: permission.NewChannelRoleAuthorizer(deps.Db),
		},
	)
}

var NewValidator = func() (protovalidate.Validator, error) {
	return protovalidate.New()
}

func BaseInterceptors() (grpc.ServerOption, error) {
	validator, err := NewValidator()

	if err != nil {
		return nil, err
	}

	return grpc.ChainUnaryInterceptor(
		middleware.RequestIdInterceptor(),
		middleware.RequestLoggerInterceptor(),
		middleware.AuthJwtInterceptor(),
		protovalidate_middleware.UnaryServerInterceptor(validator),
	), nil
}
