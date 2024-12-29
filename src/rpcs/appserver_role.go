package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_server "mist/src/protos/server/v1"
	"mist/src/service"
)

func (s *AppserverGRPCService) CreateAppserverRole(
	ctx context.Context, req *pb_server.CreateAppserverRoleRequest,
) (*pb_server.CreateAppserverRoleResponse, error) {
	asRoleService := service.NewAppserverRoleService(s.DbcPool, ctx)
	appserverSub, err := asRoleService.Create(req.GetAppserverId(), req.Name)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return response
	return &pb_server.CreateAppserverRoleResponse{
		AppserverRole: asRoleService.PgTypeToPb(appserverSub),
	}, nil
}

func (s *AppserverGRPCService) GetAllAppserverRoles(
	ctx context.Context, req *pb_server.GetAllAppserverRolesRequest,
) (*pb_server.GetAllAppserverRolesResponse, error) {
	// Initialize the service for AppserveRole
	asRoleService := service.NewAppserverRoleService(s.DbcPool, ctx)

	results, _ := asRoleService.ListAppserverRoles(req.GetAppserverId())

	// Construct the response
	response := &pb_server.GetAllAppserverRolesResponse{
		AppserverRoles: make([]*pb_server.AppserverRole, 0, len(results)),
	}

	// Convert list of AppserveRoles to protobuf
	for _, result := range results {
		response.AppserverRoles = append(response.AppserverRoles, asRoleService.PgTypeToPb(&result))
	}

	return response, nil
}

func (s *AppserverGRPCService) DeleteAppserverRole(
	ctx context.Context, req *pb_server.DeleteAppserverRoleRequest,
) (*pb_server.DeleteAppserverRoleResponse, error) {
	// Initialize the service for AppserveRole
	asRoleService := service.NewAppserverRoleService(s.DbcPool, ctx)
	jwtClaims, _ := middleware.GetJWTClaims(ctx)

	// Call delete service method
	err := asRoleService.DeleteByAppserver(req.GetId(), jwtClaims.UserID)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return success response
	return &pb_server.DeleteAppserverRoleResponse{}, nil
}
