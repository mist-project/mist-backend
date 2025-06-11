package rpcs

import (
	"context"

	"github.com/google/uuid"

	"mist/src/faults"
	"mist/src/permission"
	"mist/src/protos/v1/appserver_role"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

func (s *AppserverRoleGRPCService) Create(
	ctx context.Context, req *appserver_role.CreateRequest,
) (*appserver_role.CreateResponse, error) {

	var err error

	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId})

	if err = s.Auth.Authorize(ctx, nil, permission.ActionCreate); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	roleService := service.NewAppserverRoleService(
		ctx, &service.ServiceDeps{Db: s.Deps.Db, MProducer: s.Deps.MProducer},
	)
	aRole, err := roleService.Create(qx.CreateAppserverRoleParams{
		Name: req.Name, AppserverID: serverId, AppserverPermissionMask: req.AppserverPermissionMask,
		ChannelPermissionMask: req.ChannelPermissionMask,
	})

	// Error handling
	if err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	// Return response
	return &appserver_role.CreateResponse{
		AppserverRole: roleService.PgTypeToPb(aRole),
	}, nil
}

func (s *AppserverRoleGRPCService) ListServerRoles(
	ctx context.Context, req *appserver_role.ListServerRolesRequest,
) (*appserver_role.ListServerRolesResponse, error) {

	var (
		err error
	)
	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId})

	if err = s.Auth.Authorize(ctx, nil, permission.ActionRead); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	roleService := service.NewAppserverRoleService(
		ctx, &service.ServiceDeps{Db: s.Deps.Db, MProducer: s.Deps.MProducer},
	)
	results, err := roleService.ListAppserverRoles(serverId)

	// Error handling
	if err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	// Construct the response
	response := &appserver_role.ListServerRolesResponse{
		AppserverRoles: make([]*appserver_role.AppserverRole, 0, len(results)),
	}
	// Convert list of AppserveRoles to protobuf
	for _, result := range results {
		response.AppserverRoles = append(response.AppserverRoles, roleService.PgTypeToPb(&result))
	}

	return response, nil
}

func (s *AppserverRoleGRPCService) Delete(
	ctx context.Context, req *appserver_role.DeleteRequest,
) (*appserver_role.DeleteResponse, error) {

	var err error

	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId})

	if err = s.Auth.Authorize(ctx, &req.Id, permission.ActionDelete); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	// Initialize the service for AppserveRole
	roleId, _ := uuid.Parse(req.Id)

	// Call delete service method
	err = service.NewAppserverRoleService(
		ctx, &service.ServiceDeps{Db: s.Deps.Db, MProducer: s.Deps.MProducer},
	).Delete(roleId)

	// Error handling
	if err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	// Return success response
	return &appserver_role.DeleteResponse{}, nil
}
