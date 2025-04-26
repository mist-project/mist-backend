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
)

type Authorizer interface {
	Authorize(ctx context.Context, objId *string, action Action, subAction string) error
}
