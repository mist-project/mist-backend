package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_appserver "mist/src/protos/v1/appserver"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

func (s *AppserverGRPCService) CreateAppserver(
	ctx context.Context, req *pb_appserver.CreateAppserverRequest,
) (*pb_appserver.CreateAppserverResponse, error) {

	as := service.NewAppserverService(s.DbcPool, ctx)
	claims, _ := middleware.GetJWTClaims(ctx)
	appserver, err := service.NewAppserverService(s.DbcPool, ctx).Create(req.GetName(), claims.UserID)

	if err != nil {
		return nil, ErrorHandler(err)
	}

	service.NewAppserverSubService(s.DbcPool, ctx).Create(appserver.ID.String(), claims.UserID)

	pbA := as.PgTypeToPb(appserver)
	pbA.IsOwner = appserver.AppuserID.String() == claims.UserID
	return &pb_appserver.CreateAppserverResponse{
		Appserver: pbA,
	}, nil
}

func (s *AppserverGRPCService) GetByIdAppserver(
	ctx context.Context, req *pb_appserver.GetByIdAppserverRequest,
) (*pb_appserver.GetByIdAppserverResponse, error) {

	var (
		err       error
		appserver *qx.Appserver
	)
	claims, _ := middleware.GetJWTClaims(ctx)
	as := service.NewAppserverService(s.DbcPool, ctx)

	if appserver, err = as.GetById(req.GetId()); err != nil {
		return nil, ErrorHandler(err)
	}

	pbA := as.PgTypeToPb(appserver)
	pbA.IsOwner = appserver.AppuserID.String() == claims.UserID
	return &pb_appserver.GetByIdAppserverResponse{Appserver: pbA}, nil
}

func (s *AppserverGRPCService) ListAppservers(
	ctx context.Context, req *pb_appserver.ListAppserversRequest,
) (*pb_appserver.ListAppserversResponse, error) {
	as := service.NewAppserverService(s.DbcPool, ctx)
	claims, _ := middleware.GetJWTClaims(ctx)

	// TODO: Figure out what can go wrong to add error handler
	appservers, _ := as.List(req.GetName(), claims.UserID)

	response := &pb_appserver.ListAppserversResponse{}

	// Resize the array
	response.Appservers = make([]*pb_appserver.Appserver, 0, len(appservers))

	for _, appserver := range appservers {
		pbA := as.PgTypeToPb(&appserver)
		pbA.IsOwner = appserver.AppuserID.String() == claims.UserID
		response.Appservers = append(response.Appservers, pbA)
	}

	return response, nil
}

func (s *AppserverGRPCService) DeleteAppserver(
	ctx context.Context, req *pb_appserver.DeleteAppserverRequest,
) (*pb_appserver.DeleteAppserverResponse, error) {

	claims, _ := middleware.GetJWTClaims(ctx)

	err := service.NewAppserverService(s.DbcPool, ctx).Delete(req.GetId(), claims.UserID)

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_appserver.DeleteAppserverResponse{}, nil
}
