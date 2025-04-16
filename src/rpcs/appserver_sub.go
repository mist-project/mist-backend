package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_appserver "mist/src/protos/v1/appserver"
	"mist/src/service"
)

func (s *AppserverGRPCService) CreateAppserverSub(
	ctx context.Context, req *pb_appserver.CreateAppserverSubRequest,
) (*pb_appserver.CreateAppserverSubResponse, error) {

	// Initialize the service for AppserverSub
	ass := service.NewAppserverSubService(s.DbcPool, ctx)

	claims, _ := middleware.GetJWTClaims(ctx)

	appserverSub, err := ass.Create(req.GetAppserverId(), claims.UserID)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return response
	return &pb_appserver.CreateAppserverSubResponse{
		AppserverSub: ass.PgTypeToPb(appserverSub),
	}, nil
}

func (s *AppserverGRPCService) GetUserAppserverSubs(
	ctx context.Context, req *pb_appserver.GetUserAppserverSubsRequest,
) (*pb_appserver.GetUserAppserverSubsResponse, error) {

	// Initialize the service for AppserverSub
	ass := service.NewAppserverSubService(s.DbcPool, ctx)

	claims, _ := middleware.GetJWTClaims(ctx)

	// TODO: Handle potential errors that can happen here
	results, _ := ass.ListUserAppserverAndSub(claims.UserID)

	// Construct the response
	response := &pb_appserver.GetUserAppserverSubsResponse{
		Appservers: make([]*pb_appserver.AppserverAndSub, 0, len(results)),
	}

	// Convert list of AppserverSubs to protobuf
	for _, result := range results {
		response.Appservers = append(response.Appservers, ass.PgAppserverSubRowToPb(&result))
	}

	return response, nil
}

func (s *AppserverGRPCService) GetAllUsersAppserverSubs(
	ctx context.Context, req *pb_appserver.GetAllUsersAppserverSubsRequest,
) (*pb_appserver.GetAllUsersAppserverSubsResponse, error) {

	// Initialize the service for AppserverSub
	ass := service.NewAppserverSubService(s.DbcPool, ctx)

	results, _ := ass.ListAllUsersAppserverAndSub(req.AppserverId)

	// Construct the response
	response := &pb_appserver.GetAllUsersAppserverSubsResponse{
		Appusers: make([]*pb_appserver.AppuserAndSub, 0, len(results)),
	}

	// Convert list of AppserverSubs to protobuf
	for _, result := range results {
		response.Appusers = append(response.Appusers, ass.PgUserSubRowToPb(&result))
	}

	return response, nil
}

func (s *AppserverGRPCService) DeleteAppserverSub(
	ctx context.Context, req *pb_appserver.DeleteAppserverSubRequest,
) (*pb_appserver.DeleteAppserverSubResponse, error) {

	// Initialize the service for AppserverSub
	ass := service.NewAppserverSubService(s.DbcPool, ctx)

	// Call delete service method
	err := ass.DeleteByAppserver(req.GetId())

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return success response
	return &pb_appserver.DeleteAppserverSubResponse{}, nil
}
