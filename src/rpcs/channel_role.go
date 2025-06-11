package rpcs

import (
	"context"

	"github.com/google/uuid"

	"mist/src/faults"
	"mist/src/permission"
	"mist/src/protos/v1/channel_role"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

func (s *ChannelRoleGRPCService) Create(
	ctx context.Context, req *channel_role.CreateRequest,
) (*channel_role.CreateResponse, error) {

	var err error

	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId})

	if err = s.Auth.Authorize(ctx, nil, permission.ActionCreate); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	roleId, _ := uuid.Parse(req.AppserverRoleId)
	channelId, _ := uuid.Parse(req.ChannelId)

	roleService := service.NewChannelRoleService(
		ctx, &service.ServiceDeps{Db: s.Deps.Db, MProducer: s.Deps.MProducer},
	)
	roles, err := roleService.Create(
		qx.CreateChannelRoleParams{AppserverRoleID: roleId, AppserverID: serverId, ChannelID: channelId},
	)

	// Error handling
	if err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	// Return response
	return &channel_role.CreateResponse{
		ChannelRole: roleService.PgTypeToPb(roles),
	}, nil
}

func (s *ChannelRoleGRPCService) ListChannelRoles(
	ctx context.Context, req *channel_role.ListChannelRolesRequest,
) (*channel_role.ListChannelRolesResponse, error) {

	var (
		err error
	)
	serverId, _ := uuid.Parse(req.AppserverId)
	channelId, _ := uuid.Parse(req.ChannelId)
	ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId})

	if err = s.Auth.Authorize(ctx, nil, permission.ActionRead); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	roleService := service.NewChannelRoleService(
		ctx, &service.ServiceDeps{Db: s.Deps.Db, MProducer: s.Deps.MProducer},
	)
	results, err := roleService.ListChannelRoles(channelId)

	// Error handling
	if err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	// Construct the response
	response := &channel_role.ListChannelRolesResponse{
		ChannelRoles: make([]*channel_role.ChannelRole, 0, len(results)),
	}
	// Convert list of AppserveRoles to protobuf
	for _, result := range results {
		response.ChannelRoles = append(response.ChannelRoles, roleService.PgTypeToPb(&result))
	}

	return response, nil
}

func (s *ChannelRoleGRPCService) Delete(
	ctx context.Context, req *channel_role.DeleteRequest,
) (*channel_role.DeleteResponse, error) {

	var err error

	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId})

	if err = s.Auth.Authorize(ctx, &req.Id, permission.ActionDelete); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	// Initialize the service for AppserveRole
	roleId, _ := uuid.Parse(req.Id)

	// Call delete service method
	err = service.NewChannelRoleService(
		ctx, &service.ServiceDeps{Db: s.Deps.Db, MProducer: s.Deps.MProducer},
	).Delete(roleId)

	// Error handling
	if err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	// Return success response
	return &channel_role.DeleteResponse{}, nil
}
