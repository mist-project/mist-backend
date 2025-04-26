package rpcs

import (
	"context"

	"github.com/google/uuid"

	"mist/src/middleware"
	pb_appserverrole "mist/src/protos/v1/appserver_role"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

func (s *AppserverRoleGRPCService) CreateAppserverRole(
	ctx context.Context, req *pb_appserverrole.CreateAppserverRoleRequest,
) (*pb_appserverrole.CreateAppserverRoleResponse, error) {

	serverId, _ := uuid.Parse(req.AppserverId)
	roleService := service.NewAppserverRoleService(ctx, s.DbConn, s.Db)
	aRole, err := roleService.Create(qx.CreateAppserverRoleParams{Name: req.Name, AppserverID: serverId})

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return response
	return &pb_appserverrole.CreateAppserverRoleResponse{
		AppserverRole: roleService.PgTypeToPb(aRole),
	}, nil
}

func (s *AppserverRoleGRPCService) GetAllAppserverRoles(
	ctx context.Context, req *pb_appserverrole.GetAllAppserverRolesRequest,
) (*pb_appserverrole.GetAllAppserverRolesResponse, error) {

	// Initialize the service for AppserveRole
	roleService := service.NewAppserverRoleService(ctx, s.DbConn, s.Db)
	serverId, _ := uuid.Parse(req.AppserverId)
	results, err := roleService.ListAppserverRoles(serverId)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Construct the response
	response := &pb_appserverrole.GetAllAppserverRolesResponse{
		AppserverRoles: make([]*pb_appserverrole.AppserverRole, 0, len(results)),
	}
	// Convert list of AppserveRoles to protobuf
	for _, result := range results {
		response.AppserverRoles = append(response.AppserverRoles, roleService.PgTypeToPb(&result))
	}

	return response, nil
}

func (s *AppserverRoleGRPCService) DeleteAppserverRole(
	ctx context.Context, req *pb_appserverrole.DeleteAppserverRoleRequest,
) (*pb_appserverrole.DeleteAppserverRoleResponse, error) {

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
		return nil, ErrorHandler(err)
	}

	// Return success response
	return &pb_appserverrole.DeleteAppserverRoleResponse{}, nil
}
