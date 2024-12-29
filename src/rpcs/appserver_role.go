package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_mistbe "mist/src/protos/mistbe/v1"
	"mist/src/service"
)

func (s *Grpcserver) CreateAppserverRole(
	ctx context.Context, req *pb_mistbe.CreateAppserverRoleRequest,
) (*pb_mistbe.CreateAppserverRoleResponse, error) {
	asRoleService := service.NewAppserverRoleService(s.DbcPool, ctx)
	appserverSub, err := asRoleService.Create(req.GetAppserverId(), req.Name)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return response
	return &pb_mistbe.CreateAppserverRoleResponse{
		AppserverRole: asRoleService.PgTypeToPb(appserverSub),
	}, nil
}

func (s *Grpcserver) GetAllAppserverRoles(
	ctx context.Context, req *pb_mistbe.GetAllAppserverRolesRequest,
) (*pb_mistbe.GetAllAppserverRolesResponse, error) {
	// Initialize the service for AppserveRole
	asRoleService := service.NewAppserverRoleService(s.DbcPool, ctx)

	results, _ := asRoleService.ListAppserverRoles(req.GetAppserverId())

	// Construct the response
	response := &pb_mistbe.GetAllAppserverRolesResponse{
		AppserverRoles: make([]*pb_mistbe.AppserverRole, 0, len(results)),
	}

	// Convert list of AppserveRoles to protobuf
	for _, result := range results {
		response.AppserverRoles = append(response.AppserverRoles, asRoleService.PgTypeToPb(&result))
	}

	return response, nil
}

func (s *Grpcserver) DeleteAppserverRole(
	ctx context.Context, req *pb_mistbe.DeleteAppserverRoleRequest,
) (*pb_mistbe.DeleteAppserverRoleResponse, error) {
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
	return &pb_mistbe.DeleteAppserverRoleResponse{}, nil
}
