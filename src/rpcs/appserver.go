package rpcs

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"mist/src/faults"
	"mist/src/middleware"
	"mist/src/permission"
	"mist/src/protos/v1/appserver"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

func (s *AppserverGRPCService) Create(
	ctx context.Context, req *appserver.CreateRequest,
) (*appserver.CreateResponse, error) {

	var (
		err error
		db  db.Querier
	)

	if err = s.Auth.Authorize(ctx, nil, permission.ActionCreate); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	claims, _ := middleware.GetJWTClaims(ctx)
	userId, _ := uuid.Parse(claims.UserID)

	db, err = s.Deps.Db.Begin(ctx)

	if err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	serverS := service.NewAppserverService(
		ctx, &service.ServiceDeps{Db: db, MProducer: s.Deps.MProducer},
	)

	aserver, err := serverS.Create(qx.CreateAppserverParams{Name: req.Name, AppuserID: userId})

	if err != nil {
		if rollbackErr := db.Rollback(ctx); rollbackErr != nil {
			return nil, faults.RpcCustomErrorHandler(
				ctx, faults.DatabaseError(fmt.Sprintf("rollback error: %v | %v", rollbackErr, err.Error()), slog.LevelError),
			)
		}

		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	res := serverS.PgTypeToPb(aserver)
	res.IsOwner = aserver.AppuserID == userId

	if err = db.Commit(ctx); err != nil {
		return nil, faults.RpcCustomErrorHandler(
			ctx, faults.DatabaseError(fmt.Sprintf("commit error: %v", err.Error()), slog.LevelError),
		)
	}

	return &appserver.CreateResponse{Appserver: res}, nil
}

func (s *AppserverGRPCService) GetById(
	ctx context.Context, req *appserver.GetByIdRequest,
) (*appserver.GetByIdResponse, error) {

	var (
		err     error
		aserver *qx.Appserver
	)

	if err = s.Auth.Authorize(ctx, &req.Id, permission.ActionRead); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	claims, _ := middleware.GetJWTClaims(ctx)

	as := service.NewAppserverService(
		ctx, &service.ServiceDeps{Db: s.Deps.Db, MProducer: s.Deps.MProducer},
	)
	id, _ := uuid.Parse(req.Id)

	if aserver, err = as.GetById(id); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	pbA := as.PgTypeToPb(aserver)
	pbA.IsOwner = aserver.AppuserID.String() == claims.UserID

	return &appserver.GetByIdResponse{Appserver: pbA}, nil
}

func (s *AppserverGRPCService) List(
	ctx context.Context, req *appserver.ListRequest,
) (*appserver.ListResponse, error) {

	if err := s.Auth.Authorize(ctx, nil, permission.ActionRead); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	claims, _ := middleware.GetJWTClaims(ctx)
	userId, _ := uuid.Parse(claims.UserID)

	as := service.NewAppserverService(
		ctx, &service.ServiceDeps{Db: s.Deps.Db, MProducer: s.Deps.MProducer},
	)
	var name = pgtype.Text{Valid: false, String: ""}

	if req.Name != nil {
		name.Valid = true
		name.String = req.Name.Value
	}

	appservers, _ := as.List(qx.ListAppserversParams{Name: name, AppuserID: userId})
	response := &appserver.ListResponse{}

	// Resize the array
	response.Appservers = make([]*appserver.Appserver, 0, len(appservers))

	for _, appserver := range appservers {
		pbA := as.PgTypeToPb(&appserver)
		pbA.IsOwner = appserver.AppuserID.String() == claims.UserID
		response.Appservers = append(response.Appservers, pbA)
	}

	return response, nil
}

func (s *AppserverGRPCService) Delete(
	ctx context.Context, req *appserver.DeleteRequest,
) (*appserver.DeleteResponse, error) {

	var (
		err error
		id  uuid.UUID
	)

	if err = s.Auth.Authorize(ctx, &req.Id, permission.ActionDelete); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	id, _ = uuid.Parse(req.Id)
	err = service.NewAppserverService(
		ctx, &service.ServiceDeps{Db: s.Deps.Db, MProducer: s.Deps.MProducer},
	).Delete(id)

	if err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	return &appserver.DeleteResponse{}, nil
}
