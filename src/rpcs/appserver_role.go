package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_appserver "mist/src/protos/v1/appserver"
	"mist/src/psql_db/qx"
	"mist/src/service"

	"github.com/google/uuid"
)

func (s *AppserverGRPCService) CreateAppserverRole(
	ctx context.Context, req *pb_appserver.CreateAppserverRoleRequest,
) (*pb_appserver.CreateAppserverRoleResponse, error) {

	serverId, _ := uuid.Parse(req.AppserverId)
	roleService := service.NewAppserverRoleService(s.DbConn, ctx)
	aRole, err := roleService.Create(qx.CreateAppserverRoleParams{
		Name:        req.Name,
		AppserverID: serverId,
	})

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return response
	return &pb_appserver.CreateAppserverRoleResponse{
		AppserverRole: roleService.PgTypeToPb(aRole),
	}, nil
}

func (s *AppserverGRPCService) GetAllAppserverRoles(
	ctx context.Context, req *pb_appserver.GetAllAppserverRolesRequest,
) (*pb_appserver.GetAllAppserverRolesResponse, error) {

	// Initialize the service for AppserveRole
	roleService := service.NewAppserverRoleService(s.DbConn, ctx)
	serverId, _ := uuid.Parse(req.AppserverId)
	results, err := roleService.ListAppserverRoles(serverId)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Construct the response
	response := &pb_appserver.GetAllAppserverRolesResponse{
		AppserverRoles: make([]*pb_appserver.AppserverRole, 0, len(results)),
	}
	// Convert list of AppserveRoles to protobuf
	for _, result := range results {
		response.AppserverRoles = append(response.AppserverRoles, roleService.PgTypeToPb(&result))
	}

	return response, nil
}

func (s *AppserverGRPCService) DeleteAppserverRole(
	ctx context.Context, req *pb_appserver.DeleteAppserverRoleRequest,
) (*pb_appserver.DeleteAppserverRoleResponse, error) {

	// Initialize the service for AppserveRole
	claims, _ := middleware.GetJWTClaims(ctx)
	userId, _ := uuid.Parse(claims.UserID)
	roleId, _ := uuid.Parse(req.Id)

	// Call delete service method
	err := service.NewAppserverRoleService(s.DbConn, ctx).DeleteByAppserver(
		qx.DeleteAppserverRoleParams{AppuserID: userId, ID: roleId},
	)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return success response
	return &pb_appserver.DeleteAppserverRoleResponse{}, nil
}
