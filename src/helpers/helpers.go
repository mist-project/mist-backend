package helpers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

const RequestIdKey string = "x-request-id"

func GetRequestId(ctx context.Context) string {
	requestId, ok := ctx.Value(RequestIdKey).(string)
	if !ok || requestId == "" {
		return fmt.Sprintf("UNKNOWN-%s", uuid.NewString())
	}
	return requestId
}
