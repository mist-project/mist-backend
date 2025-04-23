package rpcs_test

import (
	"errors"
	pb_appuser "mist/src/protos/v1/appuser"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ----- RPC CreateAppuser -----
func TestCreateAppuser(t *testing.T) {
	t.Run("creates_successfully", func(t *testing.T) {
		// ARRANGE
		var count int
		ctx := setup(t, func() {})

		// ACT

		response, err := TestAppuserClient.CreateAppuser(
			ctx,
			&pb_appuser.CreateAppuserRequest{Username: "someone", Id: uuid.NewString()})

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		dbConn.QueryRow(ctx, "SELECT COUNT(*) FROM appuser").Scan(&count)
		assert.NotNil(t, response)
		assert.Equal(t, 1, count)
	})

	t.Run("invalid_arguments_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestAppuserClient.CreateAppuser(ctx, &pb_appuser.CreateAppuserRequest{})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "validation error")
	})

	t.Run("error_on_db_exists_gracefully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		mockQuerier := new(MockQuerier)
		mockQuerier.On(
			"CreateAppuser", mock.Anything, mock.Anything,
		).Return(qx.Appuser{}, errors.New("a db error"))
		svc := &rpcs.AppuserGRPCService{Db: mockQuerier}

		// ACT
		_, err := svc.CreateAppuser(ctx, &pb_appuser.CreateAppuserRequest{
			Id:       uuid.NewString(),
			Username: "boo",
		})

		// ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "a db error")

	})
}
