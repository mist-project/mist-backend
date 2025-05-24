package rpcs

import (
	"context"

	"github.com/google/uuid"

	"mist/src/errors/message"
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
	ctx = context.WithValue(
		ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId},
	)

	if err = s.Auth.Authorize(ctx, nil, permission.ActionWrite, permission.SubActionCreate); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	roleId, _ := uuid.Parse(req.AppserverRoleId)
	channelId, _ := uuid.Parse(req.ChannelId)

	roleService := service.NewChannelRoleService(ctx, s.DbConn, s.Db)
	roles, err := service.NewChannelRoleService(ctx, s.DbConn, s.Db).Create(
		qx.CreateChannelRoleParams{AppserverRoleID: roleId, AppserverID: serverId, ChannelID: channelId},
	)

	// Error handling
	if err != nil {
		return nil, message.RpcErrorHandler(err)
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

	if err = s.Auth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListChannelRoles); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	roleService := service.NewChannelRoleService(ctx, s.DbConn, s.Db)
	results, err := roleService.ListChannelRoles(channelId)

	// Error handling
	if err != nil {
		return nil, message.RpcErrorHandler(err)
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

	if err = s.Auth.Authorize(ctx, &req.Id, permission.ActionDelete, permission.SubActionDelete); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	// Initialize the service for AppserveRole
	roleId, _ := uuid.Parse(req.Id)

	// Call delete service method
	err = service.NewChannelRoleService(ctx, s.DbConn, s.Db).Delete(roleId)

	// Error handling
	if err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	// Return success response
	return &channel_role.DeleteResponse{}, nil
}
