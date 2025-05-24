package rpcs

import (
	"github.com/bufbuild/protovalidate-go"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	"mist/src/middleware"
	"mist/src/permission"
	"mist/src/producer"
	pb_appserver "mist/src/protos/v1/appserver"
	pb_appserver_permission "mist/src/protos/v1/appserver_permission"
	pb_appserver_role "mist/src/protos/v1/appserver_role"
	pb_appserver_role_sub "mist/src/protos/v1/appserver_role_sub"
	pb_appserver_sub "mist/src/protos/v1/appserver_sub"
	pb_appuser "mist/src/protos/v1/appuser"
	pb_channel "mist/src/protos/v1/channel"
	pb_channel_role "mist/src/protos/v1/channel_role"
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
	DbConn   *pgxpool.Pool
	Db       db.Querier
	Auth     permission.Authorizer
	Producer producer.MessageProducer
}

type AppserverPermissionGRPCService struct {
	pb_appserver_permission.UnimplementedAppserverPermissionServiceServer
	DbConn   *pgxpool.Pool
	Db       db.Querier
	Auth     permission.Authorizer
	Producer producer.MessageProducer
}

type AppserverSubGRPCService struct {
	pb_appserver_sub.UnimplementedAppserverSubServiceServer
	DbConn   *pgxpool.Pool
	Db       db.Querier
	Auth     permission.Authorizer
	Producer *producer.MessageProducer
}

type AppserverRoleGRPCService struct {
	pb_appserver_role.UnimplementedAppserverRoleServiceServer
	DbConn   *pgxpool.Pool
	Db       db.Querier
	Auth     permission.Authorizer
	Producer producer.MessageProducer
}

type AppserverRoleSubGRPCService struct {
	pb_appserver_role_sub.UnimplementedAppserverRoleSubServiceServer
	DbConn   *pgxpool.Pool
	Db       db.Querier
	Auth     permission.Authorizer
	Producer producer.MessageProducer
}

type ChannelGRPCService struct {
	pb_channel.UnimplementedChannelServiceServer
	DbConn   *pgxpool.Pool
	Db       db.Querier
	Auth     permission.Authorizer
	Producer producer.MessageProducer
}

type ChannelRoleGRPCService struct {
	pb_channel_role.UnimplementedChannelRoleServiceServer
	DbConn   *pgxpool.Pool
	Db       db.Querier
	Auth     permission.Authorizer
	Producer producer.MessageProducer
}

func RegisterGrpcServices(s *grpc.Server, dbConn *pgxpool.Pool, mp producer.MessageProducer) {
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
	pb_appserver.RegisterAppserverServiceServer(
		s,
		&AppserverGRPCService{
			Db:     querier,
			DbConn: dbConn,
			Auth:   permission.NewAppserverAuthorizer(dbConn, querier),
		},
	)

	// ----- APPSERVER PERMISSION-----
	pb_appserver_permission.RegisterAppserverPermissionServiceServer(
		s,
		&AppserverPermissionGRPCService{
			Db:     querier,
			DbConn: dbConn,
			Auth:   permission.NewAppserverPermissionAuthorizer(dbConn, querier)},
	)

	// ----- APPSERVER ROLE -----
	pb_appserver_role.RegisterAppserverRoleServiceServer(
		s,
		&AppserverRoleGRPCService{
			Db:     querier,
			DbConn: dbConn,
			Auth:   permission.NewAppserverRoleAuthorizer(dbConn, querier)},
	)

	// ----- APPSERVER ROLE SUB -----
	pb_appserver_role_sub.RegisterAppserverRoleSubServiceServer(
		s,
		&AppserverRoleSubGRPCService{
			Db:     querier,
			DbConn: dbConn,
			Auth:   permission.NewAppserverRoleSubAuthorizer(dbConn, querier),
		},
	)

	// ----- APPSERVER SUB -----
	pb_appserver_sub.RegisterAppserverSubServiceServer(
		s,
		&AppserverSubGRPCService{
			Db:     querier,
			DbConn: dbConn,
			Auth:   permission.NewAppserverSubAuthorizer(dbConn, querier),
		},
	)

	// ----- CHANNEL -----
	pb_channel.RegisterChannelServiceServer(
		s,
		&ChannelGRPCService{
			Db:       querier,
			DbConn:   dbConn,
			Auth:     permission.NewChannelAuthorizer(dbConn, querier),
			Producer: mp,
		})

	// ----- CHANNEL ROLE -----
	pb_channel_role.RegisterChannelRoleServiceServer(
		s,
		&ChannelRoleGRPCService{
			Db:       querier,
			DbConn:   dbConn,
			Auth:     permission.NewChannelRoleAuthorizer(dbConn, querier),
			Producer: mp,
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
