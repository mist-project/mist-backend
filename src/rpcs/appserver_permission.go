package rpcs

import (
	"context"

	"github.com/google/uuid"

	"mist/src/errors/message"
	"mist/src/permission"
	"mist/src/protos/v1/appserver_permission"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

func (s *AppserverPermissionGRPCService) Create(
	ctx context.Context,
	req *appserver_permission.CreateRequest,
) (*appserver_permission.CreateResponse, error) {
	var err error

	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(
		ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId},
	)

	if err = s.Auth.Authorize(ctx, nil, permission.ActionWrite, permission.SubActionCreate); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	service := service.NewAppserverPermissionService(ctx, s.DbConn, s.Db)
	_, err = service.Create(qx.CreateAppserverPermissionParams{
		AppserverID: serverId,
		AppuserID:   uuid.MustParse(req.AppuserId),
	})

	if err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	return &appserver_permission.CreateResponse{}, nil
}

func (s *AppserverPermissionGRPCService) ListAppserverUsers(
	ctx context.Context,
	req *appserver_permission.ListAppserverUsersRequest,
) (*appserver_permission.ListAppserverUsersResponse, error) {
	var err error

	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(
		ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId},
	)

	if err = s.Auth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListAppserverUserPermsission); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	service := service.NewAppserverPermissionService(ctx, s.DbConn, s.Db)
	results, err := service.ListAppserverPermissions(serverId)

	if err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	response := &appserver_permission.ListAppserverUsersResponse{
		AppserverPermissions: make([]*appserver_permission.AppserverPermission, 0, len(results)),
	}

	for _, result := range results {
		response.AppserverPermissions = append(response.AppserverPermissions, service.PgTypeToPb(&result))
	}

	return response, nil
}

func (s *AppserverPermissionGRPCService) Delete(
	ctx context.Context,
	req *appserver_permission.DeleteRequest,
) (*appserver_permission.DeleteResponse, error) {
	var err error

	if err = s.Auth.Authorize(ctx, &req.Id, permission.ActionDelete, permission.SubActionDelete); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	permId, _ := uuid.Parse(req.Id)

	err = service.NewAppserverPermissionService(ctx, s.DbConn, s.Db).Delete(
		permId,
	)

	if err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	return &appserver_permission.DeleteResponse{}, nil
}
