package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_servers "mist/src/protos/server/v1"
	"mist/src/service"
)

func (s *Grpcserver) CreateAppserverRole(
	ctx context.Context, req *pb_servers.CreateAppserverRoleRequest,
) (*pb_servers.CreateAppserverRoleResponse, error) {
	asRoleService := service.NewAppserverRoleService(s.DbcPool, ctx)
	appserverSub, err := asRoleService.Create(req.GetAppserverId(), req.Name)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return response
	return &pb_servers.CreateAppserverRoleResponse{
		AppserverRole: asRoleService.PgTypeToPb(appserverSub),
	}, nil
}

func (s *Grpcserver) GetAllAppserverRoles(
	ctx context.Context, req *pb_servers.GetAllAppserverRolesRequest,
) (*pb_servers.GetAllAppserverRolesResponse, error) {
	// Initialize the service for AppserveRole
	asRoleService := service.NewAppserverRoleService(s.DbcPool, ctx)

	results, _ := asRoleService.ListAppserverRoles(req.GetAppserverId())

	// Construct the response
	response := &pb_servers.GetAllAppserverRolesResponse{
		AppserverRoles: make([]*pb_servers.AppserverRole, 0, len(results)),
	}

	// Convert list of AppserveRoles to protobuf
	for _, result := range results {
		response.AppserverRoles = append(response.AppserverRoles, asRoleService.PgTypeToPb(&result))
	}

	return response, nil
}

func (s *Grpcserver) DeleteAppserverRole(
	ctx context.Context, req *pb_servers.DeleteAppserverRoleRequest,
) (*pb_servers.DeleteAppserverRoleResponse, error) {
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
	return &pb_servers.DeleteAppserverRoleResponse{}, nil
}
