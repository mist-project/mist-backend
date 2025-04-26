package rpcs

import (
	"context"

	"github.com/google/uuid"

	"mist/src/middleware"
	pb_appserverrolesub "mist/src/protos/v1/appserver_role_sub"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

func (s *AppserverRoleSubGRPCService) Create(
	ctx context.Context, req *pb_appserverrolesub.CreateRequest,
) (*pb_appserverrolesub.CreateResponse, error) {
	roleSubS := service.NewAppserverRoleSubService(ctx, s.DbConn, s.Db)

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
	return &pb_appserverrolesub.CreateResponse{
		AppserverRoleSub: roleSubS.PgTypeToPb(arSub),
	}, nil
}

func (s *AppserverRoleSubGRPCService) ListServerRoleSubs(
	ctx context.Context, req *pb_appserverrolesub.ListServerRoleSubsRequest,
) (*pb_appserverrolesub.ListServerRoleSubsResponse, error) {

	// Initialize the service for AppserveRole
	serverId, _ := uuid.Parse(req.AppserverId)
	results, _ := service.NewAppserverRoleSubService(ctx, s.DbConn, s.Db).ListServerRoleSubs(serverId)

	// Construct the response
	response := &pb_appserverrolesub.ListServerRoleSubsResponse{
		AppserverRoleSubs: make([]*pb_appserverrolesub.AppserverRoleSub, 0, len(results)),
	}

	// Convert list of AppserveRoles to protobuf
	for _, result := range results {
		response.AppserverRoleSubs = append(response.AppserverRoleSubs, &pb_appserverrolesub.AppserverRoleSub{
			Id:              result.ID.String(),
			AppserverRoleId: result.AppserverRoleID.String(),
			AppuserId:       result.AppuserID.String(),
			AppserverId:     result.AppserverID.String(),
		})
	}

	return response, nil
}

func (s *AppserverRoleSubGRPCService) Delete(
	ctx context.Context, req *pb_appserverrolesub.DeleteRequest,
) (*pb_appserverrolesub.DeleteResponse, error) {

	// Initialize the service for AppserveRole
	arss := service.NewAppserverRoleSubService(ctx, s.DbConn, s.Db)
	claims, _ := middleware.GetJWTClaims(ctx)
	userId, _ := uuid.Parse(claims.UserID)
	roleSubId, _ := uuid.Parse(req.Id)

	// Call delete service method
	err := arss.Delete(qx.DeleteAppserverRoleSubParams{
		ID:        roleSubId,
		AppuserID: userId,
	})

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return success response
	return &pb_appserverrolesub.DeleteResponse{}, nil
}
