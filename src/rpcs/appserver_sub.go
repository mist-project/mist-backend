package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_servers "mist/src/protos/server/v1"
	"mist/src/service"
)

func (s *Grpcserver) CreateAppserverSub(
	ctx context.Context, req *pb_servers.CreateAppserverSubRequest,
) (*pb_servers.CreateAppserverSubResponse, error) {
	// Initialize the service for AppserverSub
	appserverSubService := service.NewAppserverSubService(s.DbcPool, ctx)

	jwtClaims, _ := middleware.GetJWTClaims(ctx)

	appserverSub, err := appserverSubService.Create(req.GetAppserverId(), jwtClaims.UserID)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return response
	return &pb_servers.CreateAppserverSubResponse{
		AppserverSub: appserverSubService.PgTypeToPb(appserverSub),
	}, nil
}

func (s *Grpcserver) GetUserAppserverSubs(
	ctx context.Context, req *pb_servers.GetUserAppserverSubsRequest,
) (*pb_servers.GetUserAppserverSubsResponse, error) {
	// Initialize the service for AppserverSub
	appserverSubService := service.NewAppserverSubService(s.DbcPool, ctx)

	jwtClaims, _ := middleware.GetJWTClaims(ctx)

	// TODO: Handle potential errors that can happen here
	results, _ := appserverSubService.ListUserAppserverAndSub(jwtClaims.UserID)

	// Construct the response
	response := &pb_servers.GetUserAppserverSubsResponse{
		Appservers: make([]*pb_servers.AppserverAndSub, 0, len(results)),
	}

	// Convert list of AppserverSubs to protobuf
	for _, result := range results {
		response.Appservers = append(response.Appservers, appserverSubService.PgUserSubRowToPb(&result))
	}

	return response, nil
}

func (s *Grpcserver) DeleteAppserverSub(
	ctx context.Context, req *pb_servers.DeleteAppserverSubRequest,
) (*pb_servers.DeleteAppserverSubResponse, error) {
	// Initialize the service for AppserverSub
	appserverSubService := service.NewAppserverSubService(s.DbcPool, ctx)

	// Call delete service method
	err := appserverSubService.DeleteByAppserver(req.GetId())

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return success response
	return &pb_servers.DeleteAppserverSubResponse{}, nil
}
