package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_server "mist/src/protos/server/v1"
	"mist/src/service"
)

func (s *AppserverGRPCService) CreateAppserverSub(
	ctx context.Context, req *pb_server.CreateAppserverSubRequest,
) (*pb_server.CreateAppserverSubResponse, error) {
	// Initialize the service for AppserverSub
	appserverSubService := service.NewAppserverSubService(s.DbcPool, ctx)

	jwtClaims, _ := middleware.GetJWTClaims(ctx)

	appserverSub, err := appserverSubService.Create(req.GetAppserverId(), jwtClaims.UserID)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return response
	return &pb_server.CreateAppserverSubResponse{
		AppserverSub: appserverSubService.PgTypeToPb(appserverSub),
	}, nil
}

func (s *AppserverGRPCService) GetUserAppserverSubs(
	ctx context.Context, req *pb_server.GetUserAppserverSubsRequest,
) (*pb_server.GetUserAppserverSubsResponse, error) {
	// Initialize the service for AppserverSub
	appserverSubService := service.NewAppserverSubService(s.DbcPool, ctx)

	jwtClaims, _ := middleware.GetJWTClaims(ctx)

	// TODO: Handle potential errors that can happen here
	results, _ := appserverSubService.ListUserAppserverAndSub(jwtClaims.UserID)

	// Construct the response
	response := &pb_server.GetUserAppserverSubsResponse{
		Appservers: make([]*pb_server.AppserverAndSub, 0, len(results)),
	}

	// Convert list of AppserverSubs to protobuf
	for _, result := range results {
		response.Appservers = append(response.Appservers, appserverSubService.PgUserSubRowToPb(&result))
	}

	return response, nil
}

func (s *AppserverGRPCService) DeleteAppserverSub(
	ctx context.Context, req *pb_server.DeleteAppserverSubRequest,
) (*pb_server.DeleteAppserverSubResponse, error) {
	// Initialize the service for AppserverSub
	appserverSubService := service.NewAppserverSubService(s.DbcPool, ctx)

	// Call delete service method
	err := appserverSubService.DeleteByAppserver(req.GetId())

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return success response
	return &pb_server.DeleteAppserverSubResponse{}, nil
}
