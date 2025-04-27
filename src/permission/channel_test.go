package permission_test

import (
	"context"
	"fmt"
	"mist/src/middleware"
	"mist/src/permission"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
	"mist/src/testutil"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChannelAuthorizer_Authorize(t *testing.T) {
	t.Parallel()

	var (
		err         error
		channelAuth = permission.NewChannelAuthorizer(testutil.TestDbConn, db.NewQuerier(qx.New(testutil.TestDbConn)))
	)

	t.Run("ActionRead", func(t *testing.T) {
		t.Run("list-appserver-channels", func(t *testing.T) {
			t.Run("Successful:subscribed_has_access", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				userId, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
				user := testutil.TestAppuser(t, &qx.Appuser{ID: userId, Username: "foo"})
				channel := testutil.TestChannel(t, nil)
				testutil.TestAppserverSub(t, &qx.AppserverSub{AppuserID: user.ID, AppserverID: channel.AppserverID})
				idStr := channel.ID.String()

				// ACT
				err = channelAuth.Authorize(ctx, &idStr, permission.ActionRead, "list-appserver-channels")

				// ASSERT
				assert.Nil(t, err)
			})

			t.Run("Error:unsubscribed_user_not_allowed", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				appserver := testutil.TestAppserver(t, nil)
				channel := testutil.TestChannel(t, &qx.Channel{AppserverID: appserver.ID})
				idStr := channel.ID.String()

				// ACT
				err = channelAuth.Authorize(ctx, &idStr, permission.ActionRead, "list-appserver-channels")

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, "(-5) Unauthorized", err.Error())
			})

			t.Run("Error:db_error_not_allowed", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				idStr := uuid.NewString()

				mockQuerier := new(testutil.MockQuerier)
				mockQuerier.On("GetChannelById", mock.Anything, mock.Anything).Return(qx.Channel{}, nil)
				mockQuerier.On("FilterAppserverSub", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

				mockChannelAuth := permission.NewChannelAuthorizer(testutil.TestDbConn, mockQuerier)

				// ACT
				err = mockChannelAuth.Authorize(ctx, &idStr, permission.ActionRead, "list-appserver-channels")

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, "(-3) database error: db error", err.Error())
			})
		})

		t.Run("detail", func(t *testing.T) {
			t.Run("Successful:subscribed_user_has_access", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				userId, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
				user := testutil.TestAppuser(t, &qx.Appuser{ID: userId, Username: "foo"})
				channel := testutil.TestChannel(t, nil)
				testutil.TestAppserverSub(t, &qx.AppserverSub{AppuserID: user.ID, AppserverID: channel.AppserverID})
				idStr := channel.ID.String()

				// ACT
				err = channelAuth.Authorize(ctx, &idStr, permission.ActionRead, "detail")

				// ASSERT
				assert.Nil(t, err)
			})

			t.Run("Error:unsubscribed_user_not_allowed", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				channel := testutil.TestChannel(t, nil)
				idStr := channel.ID.String()

				// ACT
				err = channelAuth.Authorize(ctx, &idStr, permission.ActionRead, "detail")

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, "(-5) Unauthorized", err.Error())
			})

			t.Run("Error:db_error_returns_error", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				channel := testutil.TestChannel(t, nil)
				idStr := channel.ID.String()

				mockQuerier := new(testutil.MockQuerier)
				mockQuerier.On("GetChannelById", mock.Anything, mock.Anything).Return(*channel, nil)
				mockQuerier.On("FilterAppserverSub", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

				mockChannelAuth := permission.NewChannelAuthorizer(testutil.TestDbConn, mockQuerier)

				// ACT
				err = mockChannelAuth.Authorize(ctx, &idStr, permission.ActionRead, "detail")

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, "(-3) database error: db error", err.Error())
			})
		})
	})

	t.Run("ActionWrite", func(t *testing.T) {
		t.Run("Successful:owner_can_create_channel", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			userId, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
			user := testutil.TestAppuser(t, &qx.Appuser{ID: userId, Username: "foo"})
			appserver := testutil.TestAppserver(t, &qx.Appserver{AppuserID: user.ID})
			channel := testutil.TestChannel(t, &qx.Channel{AppserverID: appserver.ID})
			idStr := channel.ID.String()

			// ACT
			err = channelAuth.Authorize(ctx, &idStr, permission.ActionWrite, "create")

			// ASSERT
			assert.Nil(t, err)
		})

		t.Run("Successful:non_owners_cannot_create", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			channel := testutil.TestChannel(t, nil)
			idStr := channel.ID.String()

			// ACT
			err = channelAuth.Authorize(ctx, &idStr, permission.ActionWrite, "create")

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})

		t.Run("Error:db_error_returns_error", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			idStr := uuid.NewString()

			mockQuerier := new(testutil.MockQuerier)
			mockQuerier.On("GetChannelById", mock.Anything, mock.Anything).Return(qx.Channel{}, nil)
			mockQuerier.On("GetAppserverById", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

			mockChannelAuth := permission.NewChannelAuthorizer(testutil.TestDbConn, mockQuerier)

			// ACT
			err = mockChannelAuth.Authorize(ctx, &idStr, permission.ActionWrite, "create")

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-3) database error: db error", err.Error())
		})
	})

	t.Run("ActionDelete", func(t *testing.T) {
		t.Run("Successful:owner_can_delete_channel", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			userId, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
			user := testutil.TestAppuser(t, &qx.Appuser{ID: userId, Username: "foo"})
			appserver := testutil.TestAppserver(t, &qx.Appserver{AppuserID: user.ID})
			channel := testutil.TestChannel(t, &qx.Channel{AppserverID: appserver.ID})
			idStr := channel.ID.String()

			// ACT
			err = channelAuth.Authorize(ctx, &idStr, permission.ActionDelete, "")

			// ASSERT
			assert.Nil(t, err)
		})

		t.Run("Error:non_owner_cannot_delete_channel", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			channel := testutil.TestChannel(t, nil)
			idStr := channel.ID.String()

			// ACT
			err = channelAuth.Authorize(ctx, &idStr, permission.ActionDelete, "")

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})

		t.Run("Error:db_error_returns_error", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			idStr := uuid.NewString()

			mockQuerier := new(testutil.MockQuerier)
			mockQuerier.On("GetChannelById", mock.Anything, mock.Anything).Return(qx.Channel{}, nil)
			mockQuerier.On("GetAppserverById", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

			mockChannelAuth := permission.NewChannelAuthorizer(testutil.TestDbConn, mockQuerier)

			// ACT
			err = mockChannelAuth.Authorize(ctx, &idStr, permission.ActionDelete, "")

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-3) database error: db error", err.Error())
		})
	})

	t.Run("Errors", func(t *testing.T) {
		t.Run("Error:invalid_userid_in_context_errors", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			_, claims := testutil.CreateJwtToken(t, &testutil.CreateTokenParams{
				Iss:       os.Getenv("MIST_API_JWT_ISSUER"),
				Aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
				SecretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
				UserId:    "invalid",
			})
			badCtx := context.WithValue(ctx, middleware.JwtClaimsK, claims)

			// ACT
			err = channelAuth.Authorize(badCtx, nil, permission.ActionRead, "")

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-1) invalid uuid", err.Error())
		})

		t.Run("Error:invalid_object_id_uuid_errors", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			idStr := "invalid"

			// ACT
			err = channelAuth.Authorize(ctx, &idStr, permission.ActionRead, "")

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-1) invalid uuid", err.Error())
		})

		t.Run("Error:object_not_found_errors", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			idStr := uuid.NewString()

			// ACT
			err = channelAuth.Authorize(ctx, &idStr, permission.ActionRead, "")

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-2) resource not found", err.Error())
		})

		t.Run("Error:undefined_permission_defaults_to_error", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			channel := testutil.TestChannel(t, nil)
			idStr := channel.ID.String()

			// ACT
			err = channelAuth.Authorize(ctx, &idStr, permission.ActionWrite, "some-random-action")

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})
	})
}
