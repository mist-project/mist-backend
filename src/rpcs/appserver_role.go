package rpcs

import (
	"context"

	"github.com/google/uuid"

	"mist/src/errors/message"
	"mist/src/permission"
	pb_appserverrole "mist/src/protos/v1/appserver_role"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

func (s *AppserverRoleGRPCService) Create(
	ctx context.Context, req *pb_appserverrole.CreateRequest,
) (*pb_appserverrole.CreateResponse, error) {

	var err error

	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(
		ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId},
	)

	if err = s.Auth.Authorize(ctx, nil, permission.ActionWrite, permission.SubActionCreate); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	roleService := service.NewAppserverRoleService(ctx, s.DbConn, s.Db)
	aRole, err := roleService.Create(qx.CreateAppserverRoleParams{Name: req.Name, AppserverID: serverId})

	// Error handling
	if err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	// Return response
	return &pb_appserverrole.CreateResponse{
		AppserverRole: roleService.PgTypeToPb(aRole),
	}, nil
}

func (s *AppserverRoleGRPCService) ListServerRoles(
	ctx context.Context, req *pb_appserverrole.ListServerRolesRequest,
) (*pb_appserverrole.ListServerRolesResponse, error) {

	var (
		err error
	)
	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId})

	if err = s.Auth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListServerRoles); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	roleService := service.NewAppserverRoleService(ctx, s.DbConn, s.Db)
	results, err := roleService.ListAppserverRoles(serverId)

	// Error handling
	if err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	// Construct the response
	response := &pb_appserverrole.ListServerRolesResponse{
		AppserverRoles: make([]*pb_appserverrole.AppserverRole, 0, len(results)),
	}
	// Convert list of AppserveRoles to protobuf
	for _, result := range results {
		response.AppserverRoles = append(response.AppserverRoles, roleService.PgTypeToPb(&result))
	}

	return response, nil
}

func (s *AppserverRoleGRPCService) Delete(
	ctx context.Context, req *pb_appserverrole.DeleteRequest,
) (*pb_appserverrole.DeleteResponse, error) {

	var err error
	if err = s.Auth.Authorize(ctx, &req.Id, permission.ActionDelete, permission.SubActionDelete); err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	// Initialize the service for AppserveRole
	roleId, _ := uuid.Parse(req.Id)

	// Call delete service method
	err = service.NewAppserverRoleService(ctx, s.DbConn, s.Db).Delete(roleId)

	// Error handling
	if err != nil {
		return nil, message.RpcErrorHandler(err)
	}

	// Return success response
	return &pb_appserverrole.DeleteResponse{}, nil
}
