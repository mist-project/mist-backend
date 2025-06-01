package permission_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"mist/src/faults"
	"mist/src/middleware"
	"mist/src/permission"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
	"mist/src/testutil"
	"mist/src/testutil/factory"
)

func TestChannelAuthorizer_Authorize(t *testing.T) {
	var (
		err         error
		channelAuth = permission.NewChannelAuthorizer(testutil.TestDbConn, db.NewQuerier(qx.New(testutil.TestDbConn)))
	)

	t.Run("ActionRead", func(t *testing.T) {
		t.Run("Successful:subscribed_user_can_read_channels", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			tu := factory.UserAppserverSub(t)

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})

			// ACT
			err = channelAuth.Authorize(ctx, nil, permission.ActionRead)

			// ASSERT
			assert.Nil(t, err)
		})

		t.Run("Error:unsubscribed_user_cannot_read", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			sub := factory.UserAppserverUnsub(t)

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: sub.Server.ID,
			})

			// ACT
			err = channelAuth.Authorize(ctx, nil, permission.ActionRead)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "user does not have permission to manage channels")
		})
	})

	t.Run("ActionWrite", func(t *testing.T) {
		t.Run(permission.SubActionCreate, func(t *testing.T) {

			t.Run("Successful:owner_can_create_channel", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverOwner(t)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})

				// ACT
				err = channelAuth.Authorize(ctx, nil, permission.ActionCreate)

				// ASSERT
				assert.Nil(t, err)
			})

			t.Run("Successful:user_with_appserver_permission_can_create_channel", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverWithAllPermissions(t)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})

				// ACT
				err = channelAuth.Authorize(ctx, nil, permission.ActionWrite)

				// ASSERT
				assert.Nil(t, err)
			})

			t.Run("Error:subscribed_user_cannot_create_channel", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverSub(t)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})

				// ACT
				err = channelAuth.Authorize(ctx, nil, permission.ActionWrite)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
				testutil.AssertCustomErrorContains(t, err, "user does not have permission to manage channels")
			})

			t.Run("Error:unsubscribed_user_cannot_create_channel", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverUnsub(t)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})

				// ACT
				err = channelAuth.Authorize(ctx, nil, permission.ActionWrite)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
				testutil.AssertCustomErrorContains(t, err, "user does not have permission to manage channels")
			})
		})
	})

	t.Run("ActionDelete", func(t *testing.T) {
		t.Run(permission.SubActionDelete, func(t *testing.T) {

			t.Run("Successful:owner_can_delete_channel", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverOwner(t)
				channel := testutil.TestChannel(t, &qx.Channel{Name: "foo", AppserverID: tu.Server.ID}, false)
				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})
				idStr := channel.ID.String()

				// ACT
				err = channelAuth.Authorize(ctx, &idStr, permission.ActionDelete)

				// ASSERT
				assert.Nil(t, err)
			})

			t.Run("Successful:user_with_permission_role_can_delete_channel", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverWithAllPermissions(t)
				channel := testutil.TestChannel(t, &qx.Channel{Name: "foo", AppserverID: tu.Server.ID}, false)
				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})
				idStr := channel.ID.String()

				// ACT
				err = channelAuth.Authorize(ctx, &idStr, permission.ActionDelete)

				// ASSERT
				assert.Nil(t, err)
			})

			t.Run("Error:subscribed_user_without_permission_cannot_delete_channel", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverSub(t)
				channel := testutil.TestChannel(t, &qx.Channel{Name: "foo", AppserverID: tu.Server.ID}, false)
				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})
				idStr := channel.ID.String()

				// ACT
				err = channelAuth.Authorize(ctx, &idStr, permission.ActionDelete)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
				testutil.AssertCustomErrorContains(t, err, "user does not have permission to manage channels")
			})

			t.Run("Error:unsubscribed_user_cannot_delete_channel", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverUnsub(t)
				channel := testutil.TestChannel(t, &qx.Channel{Name: "foo", AppserverID: tu.Server.ID}, false)
				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})
				idStr := channel.ID.String()

				// ACT
				err = channelAuth.Authorize(ctx, &idStr, permission.ActionDelete)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
				testutil.AssertCustomErrorContains(t, err, "user does not have permission to manage channels")
			})
		})
	})

	t.Run("Errors", func(t *testing.T) {
		t.Run("Error:invalid_userid_in_context", func(t *testing.T) {
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
			err = channelAuth.Authorize(badCtx, nil, permission.ActionRead)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "invalid user id: invalid")
		})

		t.Run("Error:db_error_on_sub_check", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			mockQuerier := new(testutil.MockQuerier)
			tu := factory.UserAppserverSub(t)
			mockQuerier.On("FilterAppserverSub", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("boom"))
			mockChannelAuth := permission.NewChannelAuthorizer(testutil.TestDbConn, mockQuerier)
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})
			idStr := uuid.New().String()

			// ACT
			err = mockChannelAuth.Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			testutil.AssertCustomErrorContains(t, err, "failed to check user subscription")
		})

		t.Run("Error:db_error_on_server_search", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			mockQuerier := new(testutil.MockQuerier)
			tu := factory.UserAppserverSub(t)
			mockQuerier.On("FilterAppserverSub", mock.Anything, mock.Anything).Return([]qx.FilterAppserverSubRow{
				{
					ID:          tu.Sub.ID,
					AppserverID: tu.Server.ID,
					AppuserID:   tu.User.ID,
				},
			}, nil)
			mockQuerier.On("GetChannelById", mock.Anything, mock.Anything).Return(qx.Channel{
				ID:          tu.Sub.ID,
				AppserverID: tu.Server.ID,
				Name:        "boo",
			}, nil)
			mockQuerier.On("GetAppserverById", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("boom"))
			mockChannelAuth := permission.NewChannelAuthorizer(testutil.TestDbConn, mockQuerier)
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})
			idStr := uuid.New().String()

			// ACT
			err = mockChannelAuth.Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "database error: boom")
		})

		t.Run("Error:db_error_on_user_permission_mask", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			mockQuerier := new(testutil.MockQuerier)
			tu := factory.UserAppserverSub(t)
			mockQuerier.On("FilterAppserverSub", mock.Anything, mock.Anything).Return([]qx.FilterAppserverSubRow{
				{
					ID:          tu.Sub.ID,
					AppserverID: tu.Server.ID,
					AppuserID:   tu.User.ID,
				},
			}, nil)
			mockQuerier.On("GetChannelById", mock.Anything, mock.Anything).Return(qx.Channel{
				ID:          tu.Sub.ID,
				AppserverID: tu.Server.ID,
				Name:        "boo",
			}, nil)
			mockQuerier.On("GetAppserverById", mock.Anything, mock.Anything).Return(qx.Appserver{
				ID:        tu.Server.ID,
				AppuserID: tu.Server.AppuserID,
			}, nil)
			mockQuerier.On("GetAppuserRoles", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("boom"))
			mockChannelAuth := permission.NewChannelAuthorizer(testutil.TestDbConn, mockQuerier)
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})
			idStr := uuid.New().String()

			// ACT
			err = mockChannelAuth.Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "failed to get user permissions")
		})

		t.Run("Error:invalid_object_id_format", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			tu := factory.UserAppserverSub(t)
			testutil.TestAppserverSub(t, nil, true)
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})
			badId := "invalid"

			// ACT
			err = channelAuth.Authorize(ctx, &badId, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.ValidationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "invalid uuid")
		})

		t.Run("Error:invalid_server_id_format", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			testutil.TestAppserverSub(t, nil, true)
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, "invalid")

			// ACT
			err = channelAuth.Authorize(ctx, nil, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "invalid permission-context in context")
		})

		t.Run("Error:object_id_not_found", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			tu := factory.UserAppserverSub(t)
			testutil.TestAppserverSub(t, nil, true)
			nonExistentId := uuid.NewString()
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})

			// ACT
			err = channelAuth.Authorize(ctx, &nonExistentId, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.NotFoundMessage)
			testutil.AssertCustomErrorContains(t, err, "resource not found")
		})

		t.Run("Error:nil_object_errors", func(t *testing.T) {
			ctx := testutil.Setup(t, func() {})
			tu := factory.UserAppserverSub(t)
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})
			// ARRANGE
			var nilObj *string

			// ACT
			err = channelAuth.Authorize(ctx, nilObj, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "user does not have permission to manage channels")
		})
	})
}
