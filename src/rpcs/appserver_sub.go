package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_appserver "mist/src/protos/v1/appserver"
	"mist/src/psql_db/qx"
	"mist/src/service"

	"github.com/google/uuid"
)

func (s *AppserverGRPCService) CreateAppserverSub(
	ctx context.Context, req *pb_appserver.CreateAppserverSubRequest,
) (*pb_appserver.CreateAppserverSubResponse, error) {

	// Initialize the service for AppserverSub
	subService := service.NewAppserverSubService(s.DbConn, ctx)

	claims, _ := middleware.GetJWTClaims(ctx)
	serverId, _ := uuid.Parse(req.AppserverId)
	userId, _ := uuid.Parse(claims.UserID)
	appserverSub, err := subService.Create(
		qx.CreateAppserverSubParams{
			AppserverID: serverId,
			AppuserID:   userId,
		},
	)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return response
	return &pb_appserver.CreateAppserverSubResponse{
		AppserverSub: subService.PgTypeToPb(appserverSub),
	}, nil
}

func (s *AppserverGRPCService) GetUserAppserverSubs(
	ctx context.Context, req *pb_appserver.GetUserAppserverSubsRequest,
) (*pb_appserver.GetUserAppserverSubsResponse, error) {

	// Initialize the service for AppserverSub
	subService := service.NewAppserverSubService(s.DbConn, ctx)

	claims, _ := middleware.GetJWTClaims(ctx)

	// TODO: Handle potential errors that can happen here
	userId, _ := uuid.Parse(claims.UserID)
	results, _ := subService.ListUserAppserverAndSub(userId)

	// Construct the response
	response := &pb_appserver.GetUserAppserverSubsResponse{
		Appservers: make([]*pb_appserver.AppserverAndSub, 0, len(results)),
	}

	// Convert list of AppserverSubs to protobuf
	for _, result := range results {
		pbA := subService.PgAppserverSubRowToPb(&result)
		pbA.Appserver.IsOwner = result.AppuserID.String() == claims.UserID
		response.Appservers = append(response.Appservers, pbA)
	}

	return response, nil
}

func (s *AppserverGRPCService) GetAllUsersAppserverSubs(
	ctx context.Context, req *pb_appserver.GetAllUsersAppserverSubsRequest,
) (*pb_appserver.GetAllUsersAppserverSubsResponse, error) {

	// Initialize the service for AppserverSub
	subService := service.NewAppserverSubService(s.DbConn, ctx)
	serverId, _ := uuid.Parse((req.AppserverId))

	results, _ := subService.ListAllUsersAppserverAndSub(serverId)

	// Construct the response
	response := &pb_appserver.GetAllUsersAppserverSubsResponse{
		Appusers: make([]*pb_appserver.AppuserAndSub, 0, len(results)),
	}

	// Convert list of AppserverSubs to protobuf
	for _, result := range results {
		response.Appusers = append(response.Appusers, subService.PgUserSubRowToPb(&result))
	}

	return response, nil
}

func (s *AppserverGRPCService) DeleteAppserverSub(
	ctx context.Context, req *pb_appserver.DeleteAppserverSubRequest,
) (*pb_appserver.DeleteAppserverSubResponse, error) {

	// Initialize the service for AppserverSub
	id, _ := uuid.Parse((req.Id))
	// Call delete service method
	err := service.NewAppserverSubService(s.DbConn, ctx).DeleteByAppserver(id)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return success response
	return &pb_appserver.DeleteAppserverSubResponse{}, nil
}
