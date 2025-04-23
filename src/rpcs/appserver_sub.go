package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_appserversub "mist/src/protos/v1/appserver_sub"
	"mist/src/psql_db/qx"
	"mist/src/service"

	"github.com/google/uuid"
)

func (s *AppserverSubGRPCService) CreateAppserverSub(
	ctx context.Context, req *pb_appserversub.CreateAppserverSubRequest,
) (*pb_appserversub.CreateAppserverSubResponse, error) {

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
	return &pb_appserversub.CreateAppserverSubResponse{
		AppserverSub: subService.PgTypeToPb(appserverSub),
	}, nil
}

func (s *AppserverSubGRPCService) GetUserAppserverSubs(
	ctx context.Context, req *pb_appserversub.GetUserAppserverSubsRequest,
) (*pb_appserversub.GetUserAppserverSubsResponse, error) {

	// Initialize the service for AppserverSub
	subService := service.NewAppserverSubService(s.DbConn, ctx)

	claims, _ := middleware.GetJWTClaims(ctx)

	// TODO: Handle potential errors that can happen here
	userId, _ := uuid.Parse(claims.UserID)
	results, _ := subService.ListUserAppserverAndSub(userId)

	// Construct the response
	response := &pb_appserversub.GetUserAppserverSubsResponse{
		Appservers: make([]*pb_appserversub.AppserverAndSub, 0, len(results)),
	}

	// Convert list of AppserverSubs to protobuf
	for _, result := range results {
		pbA := subService.PgAppserverSubRowToPb(&result)
		pbA.Appserver.IsOwner = result.AppuserID.String() == claims.UserID
		response.Appservers = append(response.Appservers, pbA)
	}

	return response, nil
}

func (s *AppserverSubGRPCService) GetAllUsersAppserverSubs(
	ctx context.Context, req *pb_appserversub.GetAllUsersAppserverSubsRequest,
) (*pb_appserversub.GetAllUsersAppserverSubsResponse, error) {

	// Initialize the service for AppserverSub
	subService := service.NewAppserverSubService(s.DbConn, ctx)
	serverId, _ := uuid.Parse((req.AppserverId))

	results, _ := subService.ListAllUsersAppserverAndSub(serverId)

	// Construct the response
	response := &pb_appserversub.GetAllUsersAppserverSubsResponse{
		Appusers: make([]*pb_appserversub.AppuserAndSub, 0, len(results)),
	}

	// Convert list of AppserverSubs to protobuf
	for _, result := range results {
		response.Appusers = append(response.Appusers, subService.PgUserSubRowToPb(&result))
	}

	return response, nil
}

func (s *AppserverSubGRPCService) DeleteAppserverSub(
	ctx context.Context, req *pb_appserversub.DeleteAppserverSubRequest,
) (*pb_appserversub.DeleteAppserverSubResponse, error) {

	// Initialize the service for AppserverSub
	id, _ := uuid.Parse((req.Id))
	// Call delete service method
	err := service.NewAppserverSubService(s.DbConn, ctx).DeleteByAppserver(id)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return success response
	return &pb_appserversub.DeleteAppserverSubResponse{}, nil
}
