package rpcs

import (
	"context"

	"github.com/google/uuid"

	"mist/src/faults"
	"mist/src/permission"
	"mist/src/protos/v1/appserver_role_sub"
	"mist/src/psql_db/qx"
	"mist/src/service"
)

func (s *AppserverRoleSubGRPCService) Create(
	ctx context.Context, req *appserver_role_sub.CreateRequest,
) (*appserver_role_sub.CreateResponse, error) {

	var err error

	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId})

	if err = s.Auth.Authorize(ctx, nil, permission.ActionCreate); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	roleSubS := service.NewAppserverRoleSubService(
		ctx, &service.ServiceDeps{Db: s.Deps.Db, MProducer: s.Deps.MProducer},
	)

	// TODO: Figure out what can go wrong to add error handler
	subId, _ := uuid.Parse(req.AppserverSubId)
	roleId, _ := uuid.Parse(req.AppserverRoleId)
	userId, _ := uuid.Parse(req.AppuserId)

	arSub, err := roleSubS.Create(
		qx.CreateAppserverRoleSubParams{
			AppserverSubID:  subId,
			AppserverRoleID: roleId,
			AppuserID:       userId,
			AppserverID:     serverId,
		},
	)

	// Error handling
	if err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	// Return response
	return &appserver_role_sub.CreateResponse{
		AppserverRoleSub: roleSubS.PgTypeToPb(arSub),
	}, nil
}

func (s *AppserverRoleSubGRPCService) ListServerRoleSubs(
	ctx context.Context, req *appserver_role_sub.ListServerRoleSubsRequest,
) (*appserver_role_sub.ListServerRoleSubsResponse, error) {

	var (
		err error
	)
	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId})

	if err = s.Auth.Authorize(ctx, nil, permission.ActionRead); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	results, _ := service.NewAppserverRoleSubService(
		ctx, &service.ServiceDeps{Db: s.Deps.Db, MProducer: s.Deps.MProducer},
	).ListServerRoleSubs(serverId)

	// Construct the response
	response := &appserver_role_sub.ListServerRoleSubsResponse{
		AppserverRoleSubs: make([]*appserver_role_sub.AppserverRoleSub, 0, len(results)),
	}

	// Convert list of AppserveRoles to protobuf
	for _, result := range results {
		response.AppserverRoleSubs = append(response.AppserverRoleSubs, &appserver_role_sub.AppserverRoleSub{
			Id:              result.ID.String(),
			AppserverRoleId: result.AppserverRoleID.String(),
			AppuserId:       result.AppuserID.String(),
			AppserverId:     result.AppserverID.String(),
		})
	}

	return response, nil
}

func (s *AppserverRoleSubGRPCService) Delete(
	ctx context.Context, req *appserver_role_sub.DeleteRequest,
) (*appserver_role_sub.DeleteResponse, error) {

	var err error

	serverId, _ := uuid.Parse(req.AppserverId)
	ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{AppserverId: serverId})

	if err = s.Auth.Authorize(ctx, &req.Id, permission.ActionDelete); err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	// Initialize the service for AppserveRole
	arss := service.NewAppserverRoleSubService(
		ctx, &service.ServiceDeps{Db: s.Deps.Db, MProducer: s.Deps.MProducer},
	)
	roleSubId, _ := uuid.Parse(req.Id)

	// Call delete service method
	err = arss.Delete(roleSubId)

	// Error handling
	if err != nil {
		return nil, faults.RpcCustomErrorHandler(ctx, faults.ExtendError(err))
	}

	// Return success response
	return &appserver_role_sub.DeleteResponse{}, nil
}
