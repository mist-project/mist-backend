package rpcs

import (
	"github.com/bufbuild/protovalidate-go"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	"mist/src/middleware"
	"mist/src/permission"
	pb_appserver "mist/src/protos/v1/appserver"
	pb_appserverpermission "mist/src/protos/v1/appserver_permission"
	pb_appserverrole "mist/src/protos/v1/appserver_role"
	pb_appserverrolesub "mist/src/protos/v1/appserver_role_sub"
	pb_appserversub "mist/src/protos/v1/appserver_sub"
	pb_appuser "mist/src/protos/v1/appuser"
	pb_channel "mist/src/protos/v1/channel"
	pb_channelrole "mist/src/protos/v1/channel_role"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
)

type AppuserGRPCService struct {
	pb_appuser.UnimplementedAppuserServiceServer
	DbConn *pgxpool.Pool
	Db     db.Querier
}

type AppserverGRPCService struct {
	pb_appserver.UnimplementedAppserverServiceServer
	DbConn *pgxpool.Pool
	Db     db.Querier
	Auth   permission.Authorizer
}

type AppserverPermissionGRPCService struct {
	pb_appserverpermission.UnimplementedAppserverPermissionServiceServer
	DbConn *pgxpool.Pool
	Db     db.Querier
	Auth   permission.Authorizer
}

type AppserverSubGRPCService struct {
	pb_appserversub.UnimplementedAppserverSubServiceServer
	DbConn *pgxpool.Pool
	Db     db.Querier
	Auth   permission.Authorizer
}

type AppserverRoleGRPCService struct {
	pb_appserverrole.UnimplementedAppserverRoleServiceServer
	DbConn *pgxpool.Pool
	Db     db.Querier
	Auth   permission.Authorizer
}

type AppserverRoleSubGRPCService struct {
	pb_appserverrolesub.UnimplementedAppserverRoleSubServiceServer
	DbConn *pgxpool.Pool
	Db     db.Querier
	Auth   permission.Authorizer
}

type ChannelGRPCService struct {
	pb_channel.UnimplementedChannelServiceServer
	DbConn *pgxpool.Pool
	Db     db.Querier
	Auth   permission.Authorizer
}

type ChannelRoleGRPCService struct {
	pb_channelrole.UnimplementedChannelRoleServiceServer
	DbConn *pgxpool.Pool
	Db     db.Querier
	Auth   permission.Authorizer
}

func RegisterGrpcServices(s *grpc.Server, dbConn *pgxpool.Pool) {
	querier := db.NewQuerier(qx.New(dbConn))

	// ----- APPUSER -----
	pb_appuser.RegisterAppuserServiceServer(
		s,
		&AppuserGRPCService{
			Db:     querier,
			DbConn: dbConn,
		},
	)

	// ----- APPSERVER -----
	pb_appserver.RegisterAppserverServiceServer(s,
		&AppserverGRPCService{
			Db:     querier,
			DbConn: dbConn,
			Auth:   permission.NewAppserverAuthorizer(dbConn, querier),
		},
	)

	// ----- APPSERVER PERMISSION-----
	pb_appserverpermission.RegisterAppserverPermissionServiceServer(
		s, &AppserverPermissionGRPCService{
			Db:     querier,
			DbConn: dbConn,
			Auth:   permission.NewAppserverPermissionAuthorizer(dbConn, querier)},
	)

	// ----- APPSERVER ROLE -----
	pb_appserverrole.RegisterAppserverRoleServiceServer(s, &AppserverRoleGRPCService{
		Db:     querier,
		DbConn: dbConn,
		Auth:   permission.NewAppserverRoleAuthorizer(dbConn, querier)},
	)

	// ----- APPSERVER ROLE SUB -----
	pb_appserverrolesub.RegisterAppserverRoleSubServiceServer(
		s,
		&AppserverRoleSubGRPCService{
			Db:     querier,
			DbConn: dbConn,
			Auth:   permission.NewAppserverRoleSubAuthorizer(dbConn, querier),
		},
	)

	// ----- APPSERVER SUB -----
	pb_appserversub.RegisterAppserverSubServiceServer(
		s,
		&AppserverSubGRPCService{
			Db:     querier,
			DbConn: dbConn,
			Auth:   permission.NewAppserverSubAuthorizer(dbConn, querier),
		},
	)

	// ----- CHANNEL -----
	pb_channel.RegisterChannelServiceServer(s,
		&ChannelGRPCService{
			Db:     querier,
			DbConn: dbConn,
			Auth:   permission.NewChannelAuthorizer(dbConn, querier),
		})

	// ----- CHANNEL ROLE -----
	pb_channelrole.RegisterChannelRoleServiceServer(s,
		&ChannelRoleGRPCService{
			Db:     querier,
			DbConn: dbConn,
			Auth:   permission.NewChannelRoleAuthorizer(dbConn, querier),
		})
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
		middleware.AuthJwtInterceptor,
		protovalidate_middleware.UnaryServerInterceptor(validator),
	), nil
}
