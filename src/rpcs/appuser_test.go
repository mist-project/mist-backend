package rpcs_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"mist/src/faults"
	"mist/src/protos/v1/appuser"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
	"mist/src/testutil"
)

func TestAppuserRPCService_Create(t *testing.T) {
	t.Run("Success:creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		svc := &rpcs.AppuserGRPCService{Deps: &rpcs.GrpcDependencies{Db: db}}

		// ACT
		response, err := svc.Create(
			ctx,
			&appuser.CreateRequest{Username: "someone", Id: uuid.NewString()})

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.NotNil(t, response)
	})

	t.Run("Error:invalid_arguments_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppuserClient.Create(ctx, &appuser.CreateRequest{})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "validation error")
	})

	t.Run("Error:error_on_db_exists_gracefully", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		userId := uuid.New()
		expectedRequest := qx.CreateAppuserParams{ID: userId, Username: "boo"}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateAppuser", ctx, expectedRequest).Return(nil, fmt.Errorf("a db error"))

		svc := &rpcs.AppuserGRPCService{Deps: &rpcs.GrpcDependencies{Db: mockQuerier}}

		// ACT
		_, err := svc.Create(ctx, &appuser.CreateRequest{
			Id:       userId.String(),
			Username: "boo",
		})

		// ASSERT
		s, ok := status.FromError(err)
		assert.NotNil(t, err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, s.Code())
		assert.Contains(t, err.Error(), faults.DatabaseErrorMessage)
		mockQuerier.AssertExpectations(t)
	})
}
