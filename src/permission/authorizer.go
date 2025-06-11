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
	PermissionCtxKey string = "permission-context"
)

type Authorizer interface {
	Authorize(ctx context.Context, objId *string, action Action) error
}
