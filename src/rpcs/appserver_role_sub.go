package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_appserver "mist/src/protos/v1/appserver"
	"mist/src/service"
)

func (s *AppserverGRPCService) CreateAppserverRoleSub(
	ctx context.Context, req *pb_appserver.CreateAppserverRoleSubRequest,
) (*pb_appserver.CreateAppserverRoleSubResponse, error) {

	arss := service.NewAppserverRoleSubService(s.DbcPool, ctx)

	// TODO: Figure out what can go wrong to add error handler
	arSub, err := arss.Create(req.AppserverRoleId, req.AppserverSubId, req.AppserverId, req.AppuserId)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return response
	return &pb_appserver.CreateAppserverRoleSubResponse{
		AppserverRoleSub: arss.PgTypeToPb(arSub),
	}, nil
}

func (s *AppserverGRPCService) GetAllAppserverUserRoleSubs(
	ctx context.Context, req *pb_appserver.GetAllAppserverUserRoleSubsRequest,
) (*pb_appserver.GetAllAppserverUserRoleSubsResponse, error) {

	// Initialize the service for AppserveRole
	arss := service.NewAppserverRoleSubService(s.DbcPool, ctx)
	results, _ := arss.GetAppserverAllUserRoleSubs(req.GetAppserverId())

	// Construct the response
	response := &pb_appserver.GetAllAppserverUserRoleSubsResponse{
		AppserverRoleSubs: make([]*pb_appserver.AppserverRoleSub, 0, len(results)),
	}

	// Convert list of AppserveRoles to protobuf
	for _, result := range results {
		response.AppserverRoleSubs = append(response.AppserverRoleSubs, &pb_appserver.AppserverRoleSub{
			Id:              result.ID.String(),
			AppserverRoleId: result.AppserverRoleID.String(),
			AppuserId:       result.AppuserID.String(),
			AppserverId:     result.AppserverID.String(),
		})
	}

	return response, nil
}

func (s *AppserverGRPCService) DeleteAppserverRoleSub(
	ctx context.Context, req *pb_appserver.DeleteAppserverRoleSubRequest,
) (*pb_appserver.DeleteAppserverRoleSubResponse, error) {

	// Initialize the service for AppserveRole
	arss := service.NewAppserverRoleSubService(s.DbcPool, ctx)
	claims, _ := middleware.GetJWTClaims(ctx)

	// Call delete service method
	err := arss.DeleteRoleSub(req.GetId(), claims.UserID)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return success response
	return &pb_appserver.DeleteAppserverRoleSubResponse{}, nil
}
