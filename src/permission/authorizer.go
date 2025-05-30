package permission

import (
	"context"
)

type Action string

const (
	ActionRead   Action = "read"
	ActionCreate Action = "create"
	ActionWrite  Action = "write"
	ActionDelete Action = "delete"
)

const (
	UndefinedPermission string = "undefined permission"
	PermissionCtxKey    string = "permission-context"
)

const (
	SubActionGetById                      = "get-by-id"
	SubActionList                         = "list"
	SubActionListServerRoles              = "list-server-roles"
	SubActionListChannelRoles             = "list-channel-roles"
	SubActionListUserServerSubs           = "list-user-server-subs"
	SubActionListAppserverChannels        = "list-appserver-channels"
	SubActionListAppserverUserSubs        = "list-appserver-user-subs"
	SubActionListAppserverUserRoleSubs    = "list-appserver-user-role-subs"
	SubActionListAppserverUserPermsission = "list-appserver-user-permissions"
	SubActionCreate                       = "create"
	SubActionUpdate                       = "update"
	SubActionDelete                       = "delete"
)

type Authorizer interface {
	Authorize(ctx context.Context, objId *string, action Action) error
}
