package rpcs_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb_appuser "mist/src/protos/v1/appuser"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
	"mist/src/testutil"
)

func TestCreateAppuser(t *testing.T) {
	t.Run("Successful:creates_successfully", func(t *testing.T) {
		// ARRANGE
		var count int
		ctx := testutil.Setup(t, func() {})

		// ACT

		response, err := testutil.TestAppuserClient.CreateAppuser(
			ctx,
			&pb_appuser.CreateAppuserRequest{Username: "someone", Id: uuid.NewString()})

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		testutil.TestDbConn.QueryRow(ctx, "SELECT COUNT(*) FROM appuser").Scan(&count)
		assert.NotNil(t, response)
		assert.Equal(t, 1, count)
	})

	t.Run("Error:invalid_arguments_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestAppuserClient.CreateAppuser(ctx, &pb_appuser.CreateAppuserRequest{})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "validation error")
	})

	t.Run("Error:error_on_db_exists_gracefully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		userId := uuid.New()
		expectedRequest := qx.CreateAppuserParams{ID: userId, Username: "boo"}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On(
			"CreateAppuser", ctx, expectedRequest,
		).Return(qx.Appuser{}, fmt.Errorf("a db error"))
		svc := &rpcs.AppuserGRPCService{Db: mockQuerier, DbConn: testutil.TestDbConn}

		// ACT
		_, err := svc.CreateAppuser(ctx, &pb_appuser.CreateAppuserRequest{
			Id:       userId.String(),
			Username: "boo",
		})

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "a db error")
	})
}
