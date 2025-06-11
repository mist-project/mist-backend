package permission_test

import (
	"fmt"
	"testing"
	"time"

	"mist/src/faults"
	"mist/src/permission"
	"mist/src/psql_db/qx"
	"mist/src/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSharedAuthorizer_UserIsServerOwner(t *testing.T) {
	t.Run("Success:user_is_owner", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		userID := uuid.New()
		server := qx.Appserver{ID: uuid.New(), Name: "foo", AppuserID: userID}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppserverById", mock.Anything, server.ID).Return(server, nil)

		auth := permission.NewSharedAuthorizer(mockQuerier)

		// ACT
		isOwner, err := auth.UserIsServerOwner(ctx, userID, server.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.True(t, isOwner)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("Success:user_is_not_owner", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		userID := uuid.New()
		server := qx.Appserver{ID: uuid.New(), Name: "foo"}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppserverById", mock.Anything, server.ID).Return(server, nil)

		auth := permission.NewSharedAuthorizer(mockQuerier)

		// ACT
		isOwner, err := auth.UserIsServerOwner(ctx, userID, server.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.False(t, isOwner)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("Error:on_db_failure", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		userID := uuid.New()
		server := qx.Appserver{ID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppserverById", mock.Anything, server.ID).Return(qx.Appserver{}, fmt.Errorf("db fail"))

		auth := permission.NewSharedAuthorizer(mockQuerier)

		// ACT
		isOwner, err := auth.UserIsServerOwner(ctx, userID, server.ID)

		// ASSERT
		assert.Error(t, err)
		assert.False(t, isOwner)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: db fail")
		mockQuerier.AssertExpectations(t)
	})
}

func TestSharedAuthorizer_UserHasServerSub(t *testing.T) {
	t.Run("Success:user_has_sub", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		userID := uuid.New()
		server := qx.Appserver{ID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("FilterAppserverSub", ctx, qx.FilterAppserverSubParams{
			AppserverID: pgtype.UUID{Valid: true, Bytes: server.ID},
			AppuserID:   pgtype.UUID{Valid: true, Bytes: userID},
		}).Return([]qx.FilterAppserverSubRow{
			{
				ID:          uuid.New(),
				AppuserID:   userID,
				AppserverID: server.ID,
				CreatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
				UpdatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
			},
		}, nil)

		auth := permission.NewSharedAuthorizer(mockQuerier)

		// ACT
		hasSub, err := auth.UserHasServerSub(ctx, userID, server.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.True(t, hasSub)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("Success:user_does_not_have_sub", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		userID := uuid.New()
		server := qx.Appserver{ID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("FilterAppserverSub", mock.Anything, qx.FilterAppserverSubParams{
			AppserverID: pgtype.UUID{Valid: true, Bytes: server.ID},
			AppuserID:   pgtype.UUID{Valid: true, Bytes: userID},
		}).Return([]qx.FilterAppserverSubRow{}, nil)

		auth := permission.NewSharedAuthorizer(mockQuerier)

		// ACT
		hasSub, err := auth.UserHasServerSub(ctx, userID, server.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.False(t, hasSub)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("Error:on_db_failure", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		userID := uuid.New()
		server := qx.Appserver{ID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("FilterAppserverSub", ctx, qx.FilterAppserverSubParams{
			AppserverID: pgtype.UUID{Valid: true, Bytes: server.ID},
			AppuserID:   pgtype.UUID{Valid: true, Bytes: userID},
		}).Return(nil, fmt.Errorf("db error"))

		auth := permission.NewSharedAuthorizer(mockQuerier)

		// ACT
		hasSub, err := auth.UserHasServerSub(ctx, userID, server.ID)

		// ASSERT
		assert.Error(t, err)
		assert.False(t, hasSub)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: db error")
		mockQuerier.AssertExpectations(t)
	})
}
