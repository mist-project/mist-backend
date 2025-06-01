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

func TestAppserverRoleSubAuthorizer_Authorize(t *testing.T) {
	var (
		err         error
		roleSubAuth = permission.NewAppserverRoleSubAuthorizer(testutil.TestDbConn, db.NewQuerier(qx.New(testutil.TestDbConn)))
	)

	t.Run("ActionRead", func(t *testing.T) {
		t.Run("Successful:subscribed_user_can_list_roles", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			tu := factory.UserAppserverSub(t)

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})

			// ACT
			err = roleSubAuth.Authorize(ctx, nil, permission.ActionRead)

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
			err = roleSubAuth.Authorize(ctx, nil, permission.ActionRead)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "is not authorized to perform this action")
		})
	})
	t.Run("ActionWrite", func(t *testing.T) {
		t.Run(permission.SubActionCreate, func(t *testing.T) {

			t.Run("Successful:owner_can_create_role_sub", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverOwner(t)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})

				// ACT
				err = roleSubAuth.Authorize(ctx, nil, permission.ActionCreate)

				// ASSERT
				assert.Nil(t, err)
			})

			t.Run("Successful:user_with_permission_can_create_role_sub", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverWithAllPermissions(t)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})

				// ACT
				err = roleSubAuth.Authorize(ctx, nil, permission.ActionWrite)

				// ASSERT
				assert.Nil(t, err)
			})

			t.Run("Error:subscribed_user_cannot_create_role_sub", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverSub(t)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})

				// ACT
				err = roleSubAuth.Authorize(ctx, nil, permission.ActionWrite)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
				testutil.AssertCustomErrorContains(t, err, "is not authorized to perform this action")
			})

			t.Run("Error:unsubscribed_user_cannot_create_role_sub", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverUnsub(t)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})

				// ACT
				err = roleSubAuth.Authorize(ctx, nil, permission.ActionWrite)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
				testutil.AssertCustomErrorContains(t, err, "is not authorized to perform this action")
			})
		})
	})

	t.Run("ActionDelete", func(t *testing.T) {
		t.Run(permission.SubActionDelete, func(t *testing.T) {

			t.Run("Successful:owner_can_delete_role_sub", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				roleSub := testutil.TestAppserverRoleSub(t, nil, true)
				tu := factory.UserAppserverOwner(t)
				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})
				idStr := roleSub.ID.String()

				// ACT
				err = roleSubAuth.Authorize(ctx, &idStr, permission.ActionDelete)

				// ASSERT
				assert.Nil(t, err)
			})

			t.Run("Successful:user_with_permission_role_can_delete_role_sub", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverWithAllPermissions(t)
				user := testutil.TestAppuser(t, nil, false)
				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})
				role := testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "foo", AppserverID: tu.Server.ID}, false)
				sub := testutil.TestAppserverSub(t, &qx.AppserverSub{AppuserID: user.ID, AppserverID: tu.Server.ID}, false)
				roleSub := testutil.TestAppserverRoleSub(t, &qx.AppserverRoleSub{
					AppuserID:       user.ID,
					AppserverRoleID: role.ID,
					AppserverSubID:  sub.ID,
					AppserverID:     tu.Server.ID,
				}, false)
				idStr := roleSub.ID.String()

				// ACT
				err = roleSubAuth.Authorize(ctx, &idStr, permission.ActionDelete)

				// ASSERT
				assert.Nil(t, err)
			})

			t.Run("Error:subscribed_user_without_permission_cannot_delete_role_sub", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverSub(t)
				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})
				roleSub := testutil.TestAppserverRoleSub(t, nil, false)
				idStr := roleSub.ID.String()

				// ACT
				err = roleSubAuth.Authorize(ctx, &idStr, permission.ActionDelete)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
				testutil.AssertCustomErrorContains(t, err, "is not authorized to perform this action")
			})

			t.Run("Error:unsubscribed_user_cannot_delete_role_sub", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				tu := factory.UserAppserverUnsub(t)
				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: tu.Server.ID,
				})
				roleSub := testutil.TestAppserverRoleSub(t, nil, false)
				idStr := roleSub.ID.String()

				// ACT
				err = roleSubAuth.Authorize(ctx, &idStr, permission.ActionDelete)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
				testutil.AssertCustomErrorContains(t, err, "is not authorized to perform this action")
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
			err = roleSubAuth.Authorize(badCtx, nil, permission.ActionRead)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "invalid user id")
		})

		t.Run("Error:db_error_on_sub_check", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			mockQuerier := new(testutil.MockQuerier)
			tu := factory.UserAppserverSub(t)
			mockQuerier.On("FilterAppserverSub", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("boom"))
			mockroleSubAuth := permission.NewAppserverRoleSubAuthorizer(testutil.TestDbConn, mockQuerier)
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})
			idStr := uuid.New().String()

			// ACT
			err = mockroleSubAuth.Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "failed to check user subscription:")
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
			mockQuerier.On("GetAppserverRoleSubById", mock.Anything, mock.Anything).Return(qx.AppserverRoleSub{
				ID:          tu.Sub.ID,
				AppserverID: tu.Server.ID,
				AppuserID:   tu.User.ID,
			}, nil)
			mockQuerier.On("GetAppserverById", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("boom"))
			mockroleSubAuth := permission.NewAppserverRoleSubAuthorizer(testutil.TestDbConn, mockQuerier)
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})
			idStr := uuid.New().String()

			// ACT
			err = mockroleSubAuth.Authorize(ctx, &idStr, permission.ActionDelete)

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
			mockQuerier.On("GetAppserverRoleSubById", mock.Anything, mock.Anything).Return(qx.AppserverRoleSub{
				ID:          tu.Sub.ID,
				AppuserID:   tu.User.ID,
				AppserverID: tu.Server.ID,
			}, nil)
			mockQuerier.On("GetAppserverById", mock.Anything, mock.Anything).Return(qx.Appserver{
				ID:        tu.Server.ID,
				AppuserID: tu.Server.AppuserID,
			}, nil)
			mockQuerier.On("GetAppuserRoles", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("boom"))
			mockroleSubAuth := permission.NewAppserverRoleSubAuthorizer(testutil.TestDbConn, mockQuerier)
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})
			idStr := uuid.New().String()

			// ACT
			err = mockroleSubAuth.Authorize(ctx, &idStr, permission.ActionDelete)

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
			err = roleSubAuth.Authorize(ctx, &badId, permission.ActionDelete)

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
			err = roleSubAuth.Authorize(ctx, nil, permission.ActionDelete)

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
			err = roleSubAuth.Authorize(ctx, &nonExistentId, permission.ActionDelete)

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
			err = roleSubAuth.Authorize(ctx, nilObj, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "is not authorized to perform this action")
		})
	})
}
