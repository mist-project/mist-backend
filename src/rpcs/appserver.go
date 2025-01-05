package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_server "mist/src/protos/server/v1"
	"mist/src/service"
)

func (s *AppserverGRPCService) CreateAppserver(
	ctx context.Context, req *pb_server.CreateAppserverRequest,
) (*pb_server.CreateAppserverResponse, error) {
	appserverService := service.NewAppserverService(s.DbcPool, ctx)
	jwtClaims, _ := middleware.GetJWTClaims(ctx)

	appserver, err := service.NewAppserverService(s.DbcPool, ctx).Create(req.GetName(), jwtClaims.UserID)

	if err != nil {
		return nil, ErrorHandler(err)
	}

	service.NewAppserverSubService(s.DbcPool, ctx).Create(appserver.ID.String(), jwtClaims.UserID)

	return &pb_server.CreateAppserverResponse{
		Appserver: appserverService.PgTypeToPb(appserver),
	}, nil
}

func (s *AppserverGRPCService) GetByIdAppserver(
	ctx context.Context, req *pb_server.GetByIdAppserverRequest,
) (*pb_server.GetByIdAppserverResponse, error) {
	appserverService := service.NewAppserverService(s.DbcPool, ctx)
	appserver, err := appserverService.GetById(req.GetId())

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_server.GetByIdAppserverResponse{Appserver: appserverService.PgTypeToPb(appserver)}, nil
}

func (s *AppserverGRPCService) ListAppservers(
	ctx context.Context, req *pb_server.ListAppserversRequest,
) (*pb_server.ListAppserversResponse, error) {
	appserverService := service.NewAppserverService(s.DbcPool, ctx)
	jwtClaims, _ := middleware.GetJWTClaims(ctx)

	// TODO: Figure out what can go wrong to add error handler
	appservers, _ := appserverService.List(req.GetName(), jwtClaims.UserID)

	response := &pb_server.ListAppserversResponse{}

	// Resize the array
	response.Appservers = make([]*pb_server.Appserver, 0, len(appservers))

	for _, appserver := range appservers {
		response.Appservers = append(response.Appservers, appserverService.PgTypeToPb(&appserver))
	}

	return response, nil
}

func (s *AppserverGRPCService) DeleteAppserver(
	ctx context.Context, req *pb_server.DeleteAppserverRequest,
) (*pb_server.DeleteAppserverResponse, error) {

	jwtClaims, _ := middleware.GetJWTClaims(ctx)

	err := service.NewAppserverService(s.DbcPool, ctx).Delete(req.GetId(), jwtClaims.UserID)

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_server.DeleteAppserverResponse{}, nil
}
