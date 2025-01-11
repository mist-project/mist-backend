package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_server "mist/src/protos/server/v1"
	"mist/src/service"
)

func (s *AppserverGRPCService) CreateAppserverRoleSub(
	ctx context.Context, req *pb_server.CreateAppserverRoleSubRequest,
) (*pb_server.CreateAppserverRoleSubResponse, error) {
	asrSubService := service.NewAppserverRoleSubService(s.DbcPool, ctx)
	jwtClaims, _ := middleware.GetJWTClaims(ctx)

	// TODO: Figure out what can go wrong to add error handler
	appserverSub, err := asrSubService.Create(req.GetAppserverRoleId(), req.GetAppserverSubId(), jwtClaims.UserID)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return response
	return &pb_server.CreateAppserverRoleSubResponse{
		AppserverRoleSub: asrSubService.PgTypeToPb(appserverSub),
	}, nil
}

func (s *AppserverGRPCService) DeleteAppserverRoleSub(
	ctx context.Context, req *pb_server.DeleteAppserverRoleSubRequest,
) (*pb_server.DeleteAppserverRoleSubResponse, error) {
	// Initialize the service for AppserveRole
	asrSubService := service.NewAppserverRoleSubService(s.DbcPool, ctx)
	jwtClaims, _ := middleware.GetJWTClaims(ctx)

	// Call delete service method
	err := asrSubService.DeleteRoleSub(req.GetId(), jwtClaims.UserID)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return success response
	return &pb_server.DeleteAppserverRoleSubResponse{}, nil
}
