package rpcs

import (
	"context"

	"github.com/google/uuid"

	"mist/src/errors/message"
	"mist/src/middleware"
	pb_appserverrole "mist/src/protos/v1/appserver_role"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

func (s *AppserverRoleGRPCService) Create(
	ctx context.Context, req *pb_appserverrole.CreateRequest,
) (*pb_appserverrole.CreateResponse, error) {

	serverId, _ := uuid.Parse(req.AppserverId)
	roleService := service.NewAppserverRoleService(ctx, s.DbConn, s.Db)
	aRole, err := roleService.Create(qx.CreateAppserverRoleParams{Name: req.Name, AppserverID: serverId})

	// Error handling
	if err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	// Return response
	return &pb_appserverrole.CreateResponse{
		AppserverRole: roleService.PgTypeToPb(aRole),
	}, nil
}

func (s *AppserverRoleGRPCService) ListServerRoles(
	ctx context.Context, req *pb_appserverrole.ListServerRolesRequest,
) (*pb_appserverrole.ListServerRolesResponse, error) {

	// Initialize the service for AppserveRole
	roleService := service.NewAppserverRoleService(ctx, s.DbConn, s.Db)
	serverId, _ := uuid.Parse(req.AppserverId)
	results, err := roleService.ListAppserverRoles(serverId)

	// Error handling
	if err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	// Construct the response
	response := &pb_appserverrole.ListServerRolesResponse{
		AppserverRoles: make([]*pb_appserverrole.AppserverRole, 0, len(results)),
	}
	// Convert list of AppserveRoles to protobuf
	for _, result := range results {
		response.AppserverRoles = append(response.AppserverRoles, roleService.PgTypeToPb(&result))
	}

	return response, nil
}

func (s *AppserverRoleGRPCService) Delete(
	ctx context.Context, req *pb_appserverrole.DeleteRequest,
) (*pb_appserverrole.DeleteResponse, error) {

	// Initialize the service for AppserveRole
	claims, _ := middleware.GetJWTClaims(ctx)
	userId, _ := uuid.Parse(claims.UserID)
	roleId, _ := uuid.Parse(req.Id)

	// Call delete service method
	err := service.NewAppserverRoleService(ctx, s.DbConn, s.Db).Delete(
		qx.DeleteAppserverRoleParams{AppuserID: userId, ID: roleId},
	)

	// Error handling
	if err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	// Return success response
	return &pb_appserverrole.DeleteResponse{}, nil
}
