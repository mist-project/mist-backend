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
	"mist/src/testutil/factory"
)

func TestChannelRoleAuthorizer_Authorize(t *testing.T) {
	var (
		err             error
		channelRoleAuth = permission.NewChannelRoleAuthorizer(testutil.TestDbConn, db.NewQuerier(qx.New(testutil.TestDbConn)))
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
			err = channelRoleAuth.Authorize(ctx, nil, permission.ActionRead)

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
			err = channelRoleAuth.Authorize(ctx, nil, permission.ActionRead)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})
	})

	t.Run("ActionWrite", func(t *testing.T) {
		t.Run(permission.SubActionCreate, func(t *testing.T) {

			t.Run("Successful:owner_can_create_channel_role", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverOwner(t)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})

				// ACT
				err = channelRoleAuth.Authorize(ctx, nil, permission.ActionCreate)

				// ASSERT
				assert.Nil(t, err)
			})

			t.Run("Successful:user_with_appserver_permission_can_create_channel_role", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverWithAllPermissions(t)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})

				// ACT
				err = channelRoleAuth.Authorize(ctx, nil, permission.ActionWrite)

				// ASSERT
				assert.Nil(t, err)
			})

			t.Run("Error:subscribed_user_cannot_create_channel_role", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverSub(t)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})

				// ACT
				err = channelRoleAuth.Authorize(ctx, nil, permission.ActionWrite)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, "(-5) Unauthorized", err.Error())
			})

			t.Run("Error:unsubscribed_user_cannot_create_channel_role", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverUnsub(t)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})

				// ACT
				err = channelRoleAuth.Authorize(ctx, nil, permission.ActionWrite)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, "(-5) Unauthorized", err.Error())
			})
		})
	})

	t.Run("ActionDelete", func(t *testing.T) {
		t.Run(permission.SubActionDelete, func(t *testing.T) {

			t.Run("Successful:owner_can_delete_channel_role", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverOwner(t)
				channel := testutil.TestChannel(t, &qx.Channel{Name: "foo", AppserverID: tu.Server.ID}, false)
				role := testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "admin", AppserverID: tu.Server.ID}, false)
				channelRole := testutil.TestChannelRole(t, &qx.ChannelRole{
					AppserverID:     tu.Server.ID,
					AppserverRoleID: role.ID,
					ChannelID:       channel.ID,
				}, false)
				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})
				idStr := channelRole.ID.String()

				// ACT
				err = channelRoleAuth.Authorize(ctx, &idStr, permission.ActionDelete)

				// ASSERT
				assert.Nil(t, err)
			})

			t.Run("Successful:user_with_permission_role_can_delete_channel_role", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverWithAllPermissions(t)
				channel := testutil.TestChannel(t, &qx.Channel{Name: "foo", AppserverID: tu.Server.ID}, false)
				role := testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "anothaone", AppserverID: tu.Server.ID}, false)
				channelRole := testutil.TestChannelRole(t, &qx.ChannelRole{
					AppserverID:     tu.Server.ID,
					AppserverRoleID: role.ID,
					ChannelID:       channel.ID,
				}, false)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})
				idStr := channelRole.ID.String()

				// ACT
				err = channelRoleAuth.Authorize(ctx, &idStr, permission.ActionDelete)

				// ASSERT
				assert.Nil(t, err)
			})

			t.Run("Error:subscribed_user_without_permission_cannot_delete_channel_role", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverSub(t)
				channel := testutil.TestChannel(t, &qx.Channel{Name: "foo", AppserverID: tu.Server.ID}, false)
				role := testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "admin", AppserverID: tu.Server.ID}, false)
				channelRole := testutil.TestChannelRole(t, &qx.ChannelRole{
					AppserverID:     tu.Server.ID,
					AppserverRoleID: role.ID,
					ChannelID:       channel.ID,
				}, false)
				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})
				idStr := channelRole.ID.String()

				// ACT
				err = channelRoleAuth.Authorize(ctx, &idStr, permission.ActionDelete)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, "(-5) Unauthorized", err.Error())
			})

			t.Run("Error:unsubscribed_user_cannot_delete_channel_role", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverUnsub(t)
				channel := testutil.TestChannel(t, &qx.Channel{Name: "foo", AppserverID: tu.Server.ID}, false)
				role := testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "admin", AppserverID: tu.Server.ID}, false)
				channelRole := testutil.TestChannelRole(t, &qx.ChannelRole{
					AppserverID:     tu.Server.ID,
					AppserverRoleID: role.ID,
					ChannelID:       channel.ID,
				}, false)
				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})
				idStr := channelRole.ID.String()

				// ACT
				err = channelRoleAuth.Authorize(ctx, &idStr, permission.ActionDelete)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, "(-5) Unauthorized", err.Error())
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
			err = channelRoleAuth.Authorize(badCtx, nil, permission.ActionRead)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})

		t.Run("Error:db_error_on_sub_check", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			mockQuerier := new(testutil.MockQuerier)
			tu := factory.UserAppserverSub(t)
			mockQuerier.On("FilterAppserverSub", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("boom"))
			mockChannelRoleAuth := permission.NewChannelRoleAuthorizer(testutil.TestDbConn, mockQuerier)
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})
			idStr := uuid.New().String()

			// ACT
			err = mockChannelRoleAuth.Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
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
			mockQuerier.On("GetChannelRoleById", mock.Anything, mock.Anything).Return(qx.ChannelRole{
				ID:          tu.Sub.ID,
				AppserverID: tu.Server.ID,
			}, nil)
			mockQuerier.On("GetAppserverById", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("boom"))
			mockChannelRoleAuth := permission.NewChannelRoleAuthorizer(testutil.TestDbConn, mockQuerier)
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})
			idStr := uuid.New().String()

			// ACT
			err = mockChannelRoleAuth.Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
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
			mockQuerier.On("GetChannelRoleById", mock.Anything, mock.Anything).Return(qx.ChannelRole{
				ID:          tu.Sub.ID,
				AppserverID: tu.Server.ID,
			}, nil)
			mockQuerier.On("GetAppserverById", mock.Anything, mock.Anything).Return(qx.Appserver{
				ID:        tu.Server.ID,
				AppuserID: tu.Server.AppuserID,
			}, nil)
			mockQuerier.On("GetAppuserRoles", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("boom"))
			mockChannelRoleAuth := permission.NewChannelRoleAuthorizer(testutil.TestDbConn, mockQuerier)
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})
			idStr := uuid.New().String()

			// ACT
			err = mockChannelRoleAuth.Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
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
			err = channelRoleAuth.Authorize(ctx, &badId, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-1) invalid uuid", err.Error())
		})

		t.Run("Error:invalid_server_id_format", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			testutil.TestAppserverSub(t, nil, true)
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, "invalid")

			// ACT
			err = channelRoleAuth.Authorize(ctx, nil, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
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
			err = channelRoleAuth.Authorize(ctx, &nonExistentId, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-2) resource not found", err.Error())
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
			err = channelRoleAuth.Authorize(ctx, nilObj, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})
	})
}
