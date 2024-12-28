package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_mistbe "mist/src/protos/mistbe/v1"
	"mist/src/service"
)

func (s *Grpcserver) CreateAppserverSub(
	ctx context.Context, req *pb_mistbe.CreateAppserverSubRequest,
) (*pb_mistbe.CreateAppserverSubResponse, error) {
	// Initialize the service for AppserverSub
	appserverSubService := service.NewAppserverSubService(s.DbcPool, ctx)

	jwtClaims, _ := middleware.GetJWTClaims(ctx)

	appserverSub, err := appserverSubService.Create(req.GetAppserverId(), jwtClaims.UserID)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return response
	return &pb_mistbe.CreateAppserverSubResponse{
		AppserverSub: appserverSubService.PgTypeToPb(appserverSub),
	}, nil
}

func (s *Grpcserver) GetUserAppserverSubs(
	ctx context.Context, req *pb_mistbe.GetUserAppserverSubsRequest,
) (*pb_mistbe.GetUserAppserverSubsResponse, error) {
	// Initialize the service for AppserverSub
	appserverSubService := service.NewAppserverSubService(s.DbcPool, ctx)

	jwtClaims, _ := middleware.GetJWTClaims(ctx)

	// TODO: Handle potential errors that can happen here
	results, _ := appserverSubService.ListUserAppserverAndSub(jwtClaims.UserID)

	// Construct the response
	response := &pb_mistbe.GetUserAppserverSubsResponse{
		Appservers: make([]*pb_mistbe.AppserverAndSub, 0, len(results)),
	}

	// Convert list of AppserverSubs to protobuf
	for _, result := range results {
		response.Appservers = append(response.Appservers, appserverSubService.PgUserSubRowToPb(&result))
	}

	return response, nil
}

func (s *Grpcserver) DeleteAppserverSub(
	ctx context.Context, req *pb_mistbe.DeleteAppserverSubRequest,
) (*pb_mistbe.DeleteAppserverSubResponse, error) {
	// Initialize the service for AppserverSub
	appserverSubService := service.NewAppserverSubService(s.DbcPool, ctx)

	// Call delete service method
	err := appserverSubService.Delete(req.GetId())

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return success response
	return &pb_mistbe.DeleteAppserverSubResponse{}, nil
}
