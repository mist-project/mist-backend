package permission

import (
	"context"
)

type Action string

const (
	ActionRead   Action = "read"
	ActionWrite  Action = "write"
	ActionDelete Action = "delete"
)

const (
	UndefinedPermission string = "undefined permission"
	PermissionCtxKey    string = "permission-context"
)

const (
	SubActionGetById               = "get-by-id"
	SubActionList                  = "list"
	SubActionListServerRoles       = "list-server-roles"
	SubActionListUserServerSubs    = "list-user-server-subs"
	SubActionListAppserverUserSubs = "list-appserver-user-subs"
	SubActionCreate                = "create"
	SubActionListAppserverChannels = "list-appserver-channels"
	SubActionUpdate                = "update"
	SubActionDelete                = "delete"
)

type Authorizer interface {
	Authorize(ctx context.Context, objId *string, action Action, subAction string) error
}
