package permission_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"mist/src/middleware"
	"mist/src/permission"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
	"mist/src/testutil"
)

func TestChannelAuthorizer_Authorize(t *testing.T) {

	var (
		err         error
		channelAuth = permission.NewChannelAuthorizer(testutil.TestDbConn, db.NewQuerier(qx.New(testutil.TestDbConn)))
	)

	t.Run("ActionRead", func(t *testing.T) {
		t.Run(permission.SubActionListAppserverChannels, func(t *testing.T) {
			t.Run("Successful:subscribed_has_access", func(t *testing.T) {
				ctx := testutil.Setup(t, func() {})
				userId, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
				user := testutil.TestAppuser(t, &qx.Appuser{ID: userId, Username: "foo"}, false)
				appserver := testutil.TestAppserver(t, nil, false)
				testutil.TestAppserverSub(t, &qx.AppserverSub{AppuserID: user.ID, AppserverID: appserver.ID}, false)

				// Inject the required ctx value
				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.ChannelListAppserverChannelCtx{
					AppserverId: appserver.ID,
				})

				err = channelAuth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListAppserverChannels)

				assert.Nil(t, err)
			})

			t.Run("Error:unsubscribed_user_not_allowed", func(t *testing.T) {
				ctx := testutil.Setup(t, func() {})
				appserver := testutil.TestAppserver(t, nil, false)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.ChannelListAppserverChannelCtx{
					AppserverId: appserver.ID,
				})

				err = channelAuth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListAppserverChannels)

				assert.NotNil(t, err)
				assert.Equal(t, "(-5) Unauthorized", err.Error())
			})

			t.Run("Error:db_error_not_allowed", func(t *testing.T) {
				ctx := testutil.Setup(t, func() {})

				mockQuerier := new(testutil.MockQuerier)
				mockQuerier.On("GetChannelById", mock.Anything, mock.Anything).Return(qx.Channel{}, nil)
				mockQuerier.On("FilterAppserverSub", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

				mockChannelAuth := permission.NewChannelAuthorizer(testutil.TestDbConn, mockQuerier)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.ChannelListAppserverChannelCtx{
					AppserverId: uuid.New(),
				})

				err = mockChannelAuth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListAppserverChannels)

				assert.NotNil(t, err)
				assert.Equal(t, "(-3) database error: db error", err.Error())
			})
		})

		t.Run("get-by-id", func(t *testing.T) {
			t.Run("Successful:subscribed_user_has_access", func(t *testing.T) {
				ctx := testutil.Setup(t, func() {})
				userId, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
				user := testutil.TestAppuser(t, &qx.Appuser{ID: userId, Username: "foo"}, false)
				appserver := testutil.TestAppserver(t, nil, false)
				channel := testutil.TestChannel(t, &qx.Channel{AppserverID: appserver.ID}, false)
				testutil.TestAppserverSub(t, &qx.AppserverSub{AppuserID: user.ID, AppserverID: appserver.ID}, false)

				idStr := channel.ID.String()

				err = channelAuth.Authorize(ctx, &idStr, permission.ActionRead, "get-by-id")

				assert.Nil(t, err)
			})

			t.Run("Error:unsubscribed_user_not_allowed", func(t *testing.T) {
				ctx := testutil.Setup(t, func() {})
				channel := testutil.TestChannel(t, nil, false)
				idStr := channel.ID.String()

				err = channelAuth.Authorize(ctx, &idStr, permission.ActionRead, "get-by-id")

				assert.NotNil(t, err)
				assert.Equal(t, "(-5) Unauthorized", err.Error())
			})

			t.Run("Error:db_error_returns_error", func(t *testing.T) {
				ctx := testutil.Setup(t, func() {})
				channel := testutil.TestChannel(t, nil, false)
				idStr := channel.ID.String()

				mockQuerier := new(testutil.MockQuerier)
				mockQuerier.On("GetChannelById", mock.Anything, mock.Anything).Return(*channel, nil)
				mockQuerier.On("FilterAppserverSub", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

				mockChannelAuth := permission.NewChannelAuthorizer(testutil.TestDbConn, mockQuerier)

				err = mockChannelAuth.Authorize(ctx, &idStr, permission.ActionRead, "get-by-id")

				assert.NotNil(t, err)
				assert.Equal(t, "(-3) database error: db error", err.Error())
			})
		})
	})

	t.Run("ActionWrite", func(t *testing.T) {
		t.Run("Successful:owner_can_create_channel", func(t *testing.T) {
			ctx := testutil.Setup(t, func() {})
			userId, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
			user := testutil.TestAppuser(t, &qx.Appuser{ID: userId, Username: "foo"}, false)
			appserver := testutil.TestAppserver(t, &qx.Appserver{AppuserID: user.ID}, false)

			// Inject the required ctx value
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.ChannelCreateCtx{
				AppserverId: appserver.ID,
			})

			err = channelAuth.Authorize(ctx, nil, permission.ActionWrite, "create")

			assert.Nil(t, err)
		})

		t.Run("Error:non_owners_cannot_create", func(t *testing.T) {
			ctx := testutil.Setup(t, func() {})
			appserver := testutil.TestAppserver(t, nil, false)

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.ChannelCreateCtx{
				AppserverId: appserver.ID,
			})

			err = channelAuth.Authorize(ctx, nil, permission.ActionWrite, "create")

			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})

		t.Run("Error:db_error_returns_error", func(t *testing.T) {
			ctx := testutil.Setup(t, func() {})

			mockQuerier := new(testutil.MockQuerier)
			mockQuerier.On("GetAppserverById", mock.Anything, mock.Anything).Return(qx.Appserver{}, fmt.Errorf("db error"))

			mockChannelAuth := permission.NewChannelAuthorizer(testutil.TestDbConn, mockQuerier)

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.ChannelCreateCtx{
				AppserverId: uuid.New(),
			})

			err = mockChannelAuth.Authorize(ctx, nil, permission.ActionWrite, "create")

			assert.NotNil(t, err)
			assert.Equal(t, "(-3) database error: db error", err.Error())
		})
	})

	t.Run("ActionDelete", func(t *testing.T) {
		t.Run("Successful:owner_can_delete_channel", func(t *testing.T) {
			ctx := testutil.Setup(t, func() {})
			userId, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
			user := testutil.TestAppuser(t, &qx.Appuser{ID: userId, Username: "foo"}, false)
			appserver := testutil.TestAppserver(t, &qx.Appserver{AppuserID: user.ID}, false)
			channel := testutil.TestChannel(t, &qx.Channel{AppserverID: appserver.ID}, false)

			idStr := channel.ID.String()

			err = channelAuth.Authorize(ctx, &idStr, permission.ActionDelete, "")

			assert.Nil(t, err)
		})

		t.Run("Error:non_owner_cannot_delete_channel", func(t *testing.T) {
			ctx := testutil.Setup(t, func() {})
			channel := testutil.TestChannel(t, nil, false)

			idStr := channel.ID.String()

			err = channelAuth.Authorize(ctx, &idStr, permission.ActionDelete, "")

			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})

		t.Run("Error:db_error_returns_error", func(t *testing.T) {
			ctx := testutil.Setup(t, func() {})
			idStr := uuid.NewString()

			mockQuerier := new(testutil.MockQuerier)
			mockQuerier.On("GetChannelById", mock.Anything, mock.Anything).Return(qx.Channel{}, nil)
			mockQuerier.On("GetAppserverById", mock.Anything, mock.Anything).Return(qx.Appserver{}, fmt.Errorf("db error"))

			mockChannelAuth := permission.NewChannelAuthorizer(testutil.TestDbConn, mockQuerier)

			err = mockChannelAuth.Authorize(ctx, &idStr, permission.ActionDelete, "")

			assert.NotNil(t, err)
			assert.Equal(t, "(-3) database error: db error", err.Error())
		})
	})

	t.Run("Errors", func(t *testing.T) {
		t.Run("Error:invalid_userid_in_context_errors", func(t *testing.T) {
			ctx := testutil.Setup(t, func() {})
			_, claims := testutil.CreateJwtToken(t, &testutil.CreateTokenParams{
				Iss:       os.Getenv("MIST_API_JWT_ISSUER"),
				Aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
				SecretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
				UserId:    "invalid",
			})
			badCtx := context.WithValue(ctx, middleware.JwtClaimsK, claims)

			err = channelAuth.Authorize(badCtx, nil, permission.ActionRead, "")

			assert.NotNil(t, err)
			assert.Equal(t, "(-1) invalid uuid", err.Error())
		})

		t.Run("Error:invalid_object_id_uuid_errors", func(t *testing.T) {
			ctx := testutil.Setup(t, func() {})
			idStr := "invalid"

			err = channelAuth.Authorize(ctx, &idStr, permission.ActionRead, "")

			assert.NotNil(t, err)
			assert.Equal(t, "(-1) invalid uuid", err.Error())
		})

		t.Run("Error:object_not_found_errors", func(t *testing.T) {
			ctx := testutil.Setup(t, func() {})
			idStr := uuid.NewString()

			err = channelAuth.Authorize(ctx, &idStr, permission.ActionRead, "")

			assert.NotNil(t, err)
			assert.Equal(t, "(-2) resource not found", err.Error())
		})

		t.Run("Error:undefined_permission_defaults_to_error", func(t *testing.T) {
			ctx := testutil.Setup(t, func() {})
			channel := testutil.TestChannel(t, nil, false)
			idStr := channel.ID.String()

			err = channelAuth.Authorize(ctx, &idStr, permission.ActionWrite, "some-random-action")

			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})
	})
}
