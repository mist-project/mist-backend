package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_appserver "mist/src/protos/v1/appserver"
	"mist/src/psql_db/qx"
	"mist/src/service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (s *AppserverGRPCService) CreateAppserver(
	ctx context.Context, req *pb_appserver.CreateAppserverRequest,
) (*pb_appserver.CreateAppserverResponse, error) {

	// TODO: replace for a function to start transactions
	tx, err := s.DbConn.(*pgxpool.Pool).Begin(ctx)

	serverS := service.NewAppserverService(tx, ctx)
	claims, _ := middleware.GetJWTClaims(ctx)
	userId, _ := uuid.Parse(claims.UserID)

	appserver, err := serverS.Create(
		qx.CreateAppserverParams{
			Name:      req.Name,
			AppuserID: userId,
		},
	)

	if err != nil {
		tx.Rollback(ctx)
		return nil, ErrorHandler(err)
	}

	// once the appserver is created, add user as a subscriber
	_, err = service.NewAppserverSubService(tx, ctx).Create(
		qx.CreateAppserverSubParams{
			AppserverID: appserver.ID,
			AppuserID:   userId,
		},
	)

	if err != nil {
		tx.Rollback(ctx)
		return nil, ErrorHandler(err)
	}

	tx.Commit(ctx)
	res := serverS.PgTypeToPb(appserver)
	res.IsOwner = appserver.AppuserID == userId

	return &pb_appserver.CreateAppserverResponse{
		Appserver: res,
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
	as := service.NewAppserverService(s.DbConn, ctx)

	id, _ := uuid.Parse(req.Id)
	if appserver, err = as.GetById(id); err != nil {
		return nil, ErrorHandler(err)
	}

	pbA := as.PgTypeToPb(appserver)
	pbA.IsOwner = appserver.AppuserID.String() == claims.UserID
	return &pb_appserver.GetByIdAppserverResponse{Appserver: pbA}, nil
}

func (s *AppserverGRPCService) ListAppservers(
	ctx context.Context, req *pb_appserver.ListAppserversRequest,
) (*pb_appserver.ListAppserversResponse, error) {
	as := service.NewAppserverService(s.DbConn, ctx)
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
	id, _ := uuid.Parse(req.Id)
	userId, _ := uuid.Parse(claims.UserID)
	err := service.NewAppserverService(s.DbConn, ctx).Delete(qx.DeleteAppserverParams{
		ID:        id,
		AppuserID: userId,
	})

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_appserver.DeleteAppserverResponse{}, nil
}
