package rpcs

import (
	"github.com/bufbuild/protovalidate-go"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/jackc/pgx/v5/pgxpool"
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
	"mist/src/psql_db/qx"
)

type AppuserGRPCService struct {
	appuser.UnimplementedAppuserServiceServer
	DbConn *pgxpool.Pool
	Db     db.Querier
}

type AppserverGRPCService struct {
	appserver.UnimplementedAppserverServiceServer
	DbConn   *pgxpool.Pool
	Db       db.Querier
	Auth     permission.Authorizer
	Producer producer.MessageProducer
}

type AppserverSubGRPCService struct {
	appserver_sub.UnimplementedAppserverSubServiceServer
	DbConn   *pgxpool.Pool
	Db       db.Querier
	Auth     permission.Authorizer
	Producer producer.MessageProducer
}

type AppserverRoleGRPCService struct {
	appserver_role.UnimplementedAppserverRoleServiceServer
	DbConn   *pgxpool.Pool
	Db       db.Querier
	Auth     permission.Authorizer
	Producer producer.MessageProducer
}

type AppserverRoleSubGRPCService struct {
	appserver_role_sub.UnimplementedAppserverRoleSubServiceServer
	DbConn   *pgxpool.Pool
	Db       db.Querier
	Auth     permission.Authorizer
	Producer producer.MessageProducer
}

type ChannelGRPCService struct {
	channel.UnimplementedChannelServiceServer
	DbConn   *pgxpool.Pool
	Db       db.Querier
	Auth     permission.Authorizer
	Producer producer.MessageProducer
}

type ChannelRoleGRPCService struct {
	channel_role.UnimplementedChannelRoleServiceServer
	DbConn   *pgxpool.Pool
	Db       db.Querier
	Auth     permission.Authorizer
	Producer producer.MessageProducer
}

func RegisterGrpcServices(s *grpc.Server, dbConn *pgxpool.Pool, mp producer.MessageProducer) {
	querier := db.NewQuerier(qx.New(dbConn))

	// ----- APPUSER -----
	appuser.RegisterAppuserServiceServer(
		s,
		&AppuserGRPCService{
			Db:     querier,
			DbConn: dbConn,
		},
	)

	// ----- APPSERVER -----
	appserver.RegisterAppserverServiceServer(
		s,
		&AppserverGRPCService{
			Db:       querier,
			DbConn:   dbConn,
			Auth:     permission.NewAppserverAuthorizer(dbConn, querier),
			Producer: mp,
		},
	)

	// ----- APPSERVER ROLE -----
	appserver_role.RegisterAppserverRoleServiceServer(
		s,
		&AppserverRoleGRPCService{
			Db:       querier,
			DbConn:   dbConn,
			Auth:     permission.NewAppserverRoleAuthorizer(dbConn, querier),
			Producer: mp,
		},
	)

	// ----- APPSERVER ROLE SUB -----
	appserver_role_sub.RegisterAppserverRoleSubServiceServer(
		s,
		&AppserverRoleSubGRPCService{
			Db:       querier,
			DbConn:   dbConn,
			Auth:     permission.NewAppserverRoleSubAuthorizer(dbConn, querier),
			Producer: mp,
		},
	)

	// ----- APPSERVER SUB -----
	appserver_sub.RegisterAppserverSubServiceServer(
		s,
		&AppserverSubGRPCService{
			Db:       querier,
			DbConn:   dbConn,
			Auth:     permission.NewAppserverSubAuthorizer(dbConn, querier),
			Producer: mp,
		},
	)

	// ----- CHANNEL -----
	channel.RegisterChannelServiceServer(
		s,
		&ChannelGRPCService{
			Db:       querier,
			DbConn:   dbConn,
			Auth:     permission.NewChannelAuthorizer(dbConn, querier),
			Producer: mp,
		},
	)

	// ----- CHANNEL ROLE -----
	channel_role.RegisterChannelRoleServiceServer(
		s,
		&ChannelRoleGRPCService{
			Db:       querier,
			DbConn:   dbConn,
			Auth:     permission.NewChannelRoleAuthorizer(dbConn, querier),
			Producer: mp,
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
