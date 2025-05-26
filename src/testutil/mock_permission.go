package testutil

import (
	"context"

	"github.com/stretchr/testify/mock"

	"mist/src/permission"
)

type MockAuthorizer struct {
	mock.Mock
}

func (m *MockAuthorizer) Authorize(ctx context.Context, objId *string, action permission.Action) error {
	args := m.Called(ctx, objId, action)
	if err, ok := args.Get(0).(error); ok {
		return err
	}
	return nil // default to nil (no error) if it's not an error
}
