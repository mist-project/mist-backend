package rpcs

import (
	"context"

	"github.com/google/uuid"

	"mist/src/errors/message"
	"mist/src/middleware"
	"mist/src/permission"
	pb_appserver_sub "mist/src/protos/v1/appserver_sub"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

func (s *AppserverSubGRPCService) Create(
	ctx context.Context, req *pb_appserver_sub.CreateRequest,
) (*pb_appserver_sub.CreateResponse, error) {

	subService := service.NewAppserverSubService(ctx, s.DbConn, s.Db)
	claims, _ := middleware.GetJWTClaims(ctx)

	serverId, _ := uuid.Parse(req.AppserverId)
	userId, _ := uuid.Parse(claims.UserID)

	appserverSub, err := subService.Create(qx.CreateAppserverSubParams{AppserverID: serverId, AppuserID: userId})

	if err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	// Return response
	return &pb_appserver_sub.CreateResponse{
		AppserverSub: subService.PgTypeToPb(appserverSub),
	}, nil
}

func (s *AppserverSubGRPCService) ListUserServerSubs(
	ctx context.Context, req *pb_appserver_sub.ListUserServerSubsRequest,
) (*pb_appserver_sub.ListUserServerSubsResponse, error) {

	var err error
	if err = s.Auth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListUserServerSubs); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	// Initialize the service for AppserverSub
	subService := service.NewAppserverSubService(ctx, s.DbConn, s.Db)

	claims, _ := middleware.GetJWTClaims(ctx)

	// TODO: Handle potential errors that can happen here
	userId, _ := uuid.Parse(claims.UserID)
	results, _ := subService.ListUserServerSubs(userId)

	// Construct the response
	response := &pb_appserver_sub.ListUserServerSubsResponse{
		Appservers: make([]*pb_appserver_sub.AppserverAndSub, 0, len(results)),
	}

	// Convert list of AppserverSubs to protobuf
	for _, result := range results {
		pbA := subService.PgAppserverSubRowToPb(&result)
		pbA.Appserver.IsOwner = result.AppuserID.String() == claims.UserID
		response.Appservers = append(response.Appservers, pbA)
	}

	return response, nil
}

func (s *AppserverSubGRPCService) ListAppserverUserSubs(
	ctx context.Context, req *pb_appserver_sub.ListAppserverUserSubsRequest,
) (*pb_appserver_sub.ListAppserverUserSubsResponse, error) {

	var err error
	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId})

	if err = s.Auth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListAppserverUserSubs); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	// Initialize the service for AppserverSub
	subService := service.NewAppserverSubService(ctx, s.DbConn, s.Db)
	results, _ := subService.ListAppserverUserSubs(serverId)

	// Construct the response
	response := &pb_appserver_sub.ListAppserverUserSubsResponse{
		Appusers: make([]*pb_appserver_sub.AppuserAndSub, 0, len(results)),
	}

	// Convert list of AppserverSubs to protobuf
	for _, result := range results {
		response.Appusers = append(response.Appusers, subService.PgUserSubRowToPb(&result))
	}

	return response, nil
}

func (s *AppserverSubGRPCService) Delete(
	ctx context.Context, req *pb_appserver_sub.DeleteRequest,
) (*pb_appserver_sub.DeleteResponse, error) {

	var err error
	if err = s.Auth.Authorize(ctx, &req.Id, permission.ActionDelete, permission.SubActionDelete); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	id, _ := uuid.Parse((req.Id))
	err = service.NewAppserverSubService(ctx, s.DbConn, s.Db).Delete(id)

	// Error handling
	if err != nil {

		return nil, message.RpcErrorHandler(err)
	}

	// Return success response
	return &pb_appserver_sub.DeleteResponse{}, nil
}
