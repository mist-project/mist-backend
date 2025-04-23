package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_appserver "mist/src/protos/v1/appserver"
	"mist/src/psql_db/qx"
	"mist/src/service"

	"github.com/google/uuid"
)

func (s *AppserverGRPCService) CreateAppserverRoleSub(
	ctx context.Context, req *pb_appserver.CreateAppserverRoleSubRequest,
) (*pb_appserver.CreateAppserverRoleSubResponse, error) {

	roleSubS := service.NewAppserverRoleSubService(s.DbConn, ctx)

	// TODO: Figure out what can go wrong to add error handler
	subId, _ := uuid.Parse(req.AppserverSubId)
	roleId, _ := uuid.Parse(req.AppserverRoleId)
	userId, _ := uuid.Parse(req.AppuserId)
	serverId, _ := uuid.Parse(req.AppserverId)

	arSub, err := roleSubS.Create(
		qx.CreateAppserverRoleSubParams{
			AppserverSubID:  subId,
			AppserverRoleID: roleId,
			AppuserID:       userId,
			AppserverID:     serverId,
		},
	)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return response
	return &pb_appserver.CreateAppserverRoleSubResponse{
		AppserverRoleSub: roleSubS.PgTypeToPb(arSub),
	}, nil
}

func (s *AppserverGRPCService) GetAllAppserverUserRoleSubs(
	ctx context.Context, req *pb_appserver.GetAllAppserverUserRoleSubsRequest,
) (*pb_appserver.GetAllAppserverUserRoleSubsResponse, error) {

	// Initialize the service for AppserveRole
	serverId, _ := uuid.Parse(req.AppserverId)
	results, _ := service.NewAppserverRoleSubService(s.DbConn, ctx).GetAppserverAllUserRoleSubs(serverId)

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
	arss := service.NewAppserverRoleSubService(s.DbConn, ctx)
	claims, _ := middleware.GetJWTClaims(ctx)
	userId, _ := uuid.Parse(claims.UserID)
	roleSubId, _ := uuid.Parse(req.Id)

	// Call delete service method
	err := arss.DeleteRoleSub(qx.DeleteAppserverRoleSubParams{
		ID:        roleSubId,
		AppuserID: userId,
	})

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return success response
	return &pb_appserver.DeleteAppserverRoleSubResponse{}, nil
}
