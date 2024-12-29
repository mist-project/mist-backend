package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_servers "mist/src/protos/server/v1"
	"mist/src/service"
)

func (s *Grpcserver) CreateAppserver(
	ctx context.Context, req *pb_servers.CreateAppserverRequest,
) (*pb_servers.CreateAppserverResponse, error) {
	appserverService := service.NewAppserverService(s.DbcPool, ctx)
	jwtClaims, _ := middleware.GetJWTClaims(ctx)

	appserver, err := service.NewAppserverService(s.DbcPool, ctx).Create(req.GetName(), jwtClaims.UserID)

	if err != nil {
		return nil, ErrorHandler(err)
	}

	service.NewAppserverSubService(s.DbcPool, ctx).Create(appserver.ID.String(), jwtClaims.UserID)

	return &pb_servers.CreateAppserverResponse{
		Appserver: appserverService.PgTypeToPb(appserver),
	}, nil
}

func (s *Grpcserver) GetByIdAppserver(
	ctx context.Context, req *pb_servers.GetByIdAppserverRequest,
) (*pb_servers.GetByIdAppserverResponse, error) {
	appserverService := service.NewAppserverService(s.DbcPool, ctx)
	appserver, err := appserverService.GetById(req.GetId())

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_servers.GetByIdAppserverResponse{Appserver: appserverService.PgTypeToPb(appserver)}, nil
}

func (s *Grpcserver) ListAppservers(
	ctx context.Context, req *pb_servers.ListAppserversRequest,
) (*pb_servers.ListAppserversResponse, error) {
	appserverService := service.NewAppserverService(s.DbcPool, ctx)
	// TODO: Figure out what can go wrong to add error handler
	appservers, _ := appserverService.List(req.GetName())

	response := &pb_servers.ListAppserversResponse{}

	// Resize the array
	response.Appservers = make([]*pb_servers.Appserver, 0, len(appservers))

	for _, appserver := range appservers {
		response.Appservers = append(response.Appservers, appserverService.PgTypeToPb(&appserver))
	}

	return response, nil
}

func (s *Grpcserver) DeleteAppserver(
	ctx context.Context, req *pb_servers.DeleteAppserverRequest,
) (*pb_servers.DeleteAppserverResponse, error) {

	jwtClaims, _ := middleware.GetJWTClaims(ctx)

	err := service.NewAppserverService(s.DbcPool, ctx).Delete(req.GetId(), jwtClaims.UserID)

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_servers.DeleteAppserverResponse{}, nil
}
