package service_test

import (
	"context"
	"fmt"
	pb_appuser "mist/src/protos/v1/appuser"
	"mist/src/psql_db/qx"
	"mist/src/service"
	"mist/src/testutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAppuserService_PgTypeToPb(t *testing.T) {
	// ARRANGE
	ctx := context.Background()
	svc := service.NewAppuserService(ctx, testutil.TestDbConn, new(testutil.MockQuerier))

	id := uuid.New()
	now := time.Now()

	appuser := &qx.Appuser{ID: id, Username: "testuser", CreatedAt: pgtype.Timestamp{Time: now, Valid: true}}
	expected := &pb_appuser.Appuser{Id: id.String(), Username: "testuser", CreatedAt: timestamppb.New(now)}

	// ACT
	result := svc.PgTypeToPb(appuser)

	// ASSERT
	assert.Equal(t, expected, result)
}

func TestAppuserService_Create(t *testing.T) {
	t.Run("Successful:can_create_user", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		expectedUser := qx.Appuser{ID: uuid.New(), Username: "testuser"}
		params := qx.CreateAppuserParams{Username: expectedUser.Username}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateAppuser", ctx, params).Return(expectedUser, nil)

		svc := service.NewAppuserService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		result, err := svc.Create(params)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, result.ID)
		assert.Equal(t, expectedUser.Username, result.Username)
	})

	t.Run("Error:failure_on_db_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		params := qx.CreateAppuserParams{Username: "baduser"}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("CreateAppuser", ctx, params).Return(qx.Appuser{}, fmt.Errorf("db error"))

		svc := service.NewAppuserService(ctx, testutil.TestDbConn, mockQuerier)

		// ACT
		result, err := svc.Create(params)

		// ASSERT
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "(-3) create appuser")
	})
}
