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

	arss := service.NewAppserverRoleSubService(s.DbcPool, ctx)
	claims, _ := middleware.GetJWTClaims(ctx)

	// TODO: Figure out what can go wrong to add error handler
	arSub, err := arss.Create(req.GetAppserverRoleId(), req.GetAppserverSubId(), claims.UserID)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return response
	return &pb_server.CreateAppserverRoleSubResponse{
		AppserverRoleSub: arss.PgTypeToPb(arSub),
	}, nil
}

func (s *AppserverGRPCService) DeleteAppserverRoleSub(
	ctx context.Context, req *pb_server.DeleteAppserverRoleSubRequest,
) (*pb_server.DeleteAppserverRoleSubResponse, error) {

	// Initialize the service for AppserveRole
	arss := service.NewAppserverRoleSubService(s.DbcPool, ctx)
	claims, _ := middleware.GetJWTClaims(ctx)

	// Call delete service method
	err := arss.DeleteRoleSub(req.GetId(), claims.UserID)

	// Error handling
	if err != nil {
		return nil, ErrorHandler(err)
	}

	// Return success response
	return &pb_server.DeleteAppserverRoleSubResponse{}, nil
}
