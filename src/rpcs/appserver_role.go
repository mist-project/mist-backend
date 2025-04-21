package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_appserver "mist/src/protos/v1/appserver"
	"mist/src/service"
)

func (s *AppserverGRPCService) CreateAppserverRole(
	ctx context.Context, req *pb_appserver.CreateAppserverRoleRequest,
) (*pb_appserver.CreateAppserverRoleResponse, error) {

	ars := service.NewAppserverRoleService(s.DbcPool, ctx)
	aRole, err := ars.Create(req.GetAppserverId(), req.Name)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return response
	return &pb_appserver.CreateAppserverRoleResponse{
		AppserverRole: ars.PgTypeToPb(aRole),
	}, nil
}

func (s *AppserverGRPCService) GetAllAppserverRoles(
	ctx context.Context, req *pb_appserver.GetAllAppserverRolesRequest,
) (*pb_appserver.GetAllAppserverRolesResponse, error) {

	// Initialize the service for AppserveRole
	ars := service.NewAppserverRoleService(s.DbcPool, ctx)
	results, _ := ars.ListAppserverRoles(req.GetAppserverId())

	// Construct the response
	response := &pb_appserver.GetAllAppserverRolesResponse{
		AppserverRoles: make([]*pb_appserver.AppserverRole, 0, len(results)),
	}
	// Convert list of AppserveRoles to protobuf
	for _, result := range results {
		response.AppserverRoles = append(response.AppserverRoles, ars.PgTypeToPb(&result))
	}

	return response, nil
}

func (s *AppserverGRPCService) DeleteAppserverRole(
	ctx context.Context, req *pb_appserver.DeleteAppserverRoleRequest,
) (*pb_appserver.DeleteAppserverRoleResponse, error) {

	// Initialize the service for AppserveRole
	ars := service.NewAppserverRoleService(s.DbcPool, ctx)
	claims, _ := middleware.GetJWTClaims(ctx)

	// Call delete service method
	err := ars.DeleteByAppserver(req.GetId(), claims.UserID)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return success response
	return &pb_appserver.DeleteAppserverRoleResponse{}, nil
}
