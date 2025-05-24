package rpcs

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"mist/src/errors/message"
	"mist/src/middleware"
	"mist/src/permission"
	"mist/src/protos/v1/appserver"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

func (s *AppserverGRPCService) Create(
	ctx context.Context, req *appserver.CreateRequest,
) (*appserver.CreateResponse, error) {

	var err error

	if err = s.Auth.Authorize(ctx, nil, permission.ActionWrite, permission.SubActionCreate); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	claims, _ := middleware.GetJWTClaims(ctx)
	userId, _ := uuid.Parse(claims.UserID)

	serverS := service.NewAppserverService(ctx, s.DbConn, s.Db)
	aserver, err := serverS.Create(qx.CreateAppserverParams{Name: req.Name, AppuserID: userId})

	if err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	res := serverS.PgTypeToPb(aserver)
	res.IsOwner = aserver.AppuserID == userId

	return &appserver.CreateResponse{Appserver: res}, nil
}

func (s *AppserverGRPCService) GetById(
	ctx context.Context, req *appserver.GetByIdRequest,
) (*appserver.GetByIdResponse, error) {

	var (
		err     error
		aserver *qx.Appserver
	)

	if err = s.Auth.Authorize(ctx, &req.Id, permission.ActionRead, permission.SubActionGetById); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	claims, _ := middleware.GetJWTClaims(ctx)

	as := service.NewAppserverService(ctx, s.DbConn, s.Db)
	id, _ := uuid.Parse(req.Id)

	if aserver, err = as.GetById(id); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	pbA := as.PgTypeToPb(aserver)
	pbA.IsOwner = aserver.AppuserID.String() == claims.UserID

	return &appserver.GetByIdResponse{Appserver: pbA}, nil
}

func (s *AppserverGRPCService) List(
	ctx context.Context, req *appserver.ListRequest,
) (*appserver.ListResponse, error) {

	if err := s.Auth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionList); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	claims, _ := middleware.GetJWTClaims(ctx)
	userId, _ := uuid.Parse(claims.UserID)

	as := service.NewAppserverService(ctx, s.DbConn, s.Db)
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

	if err = s.Auth.Authorize(ctx, &req.Id, permission.ActionDelete, permission.SubActionDelete); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	id, _ = uuid.Parse(req.Id)
	err = service.NewAppserverService(ctx, s.DbConn, s.Db).Delete(id)

	if err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	return &appserver.DeleteResponse{}, nil
}
