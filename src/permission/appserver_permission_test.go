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

func TestAppserverPermissionAuthorizer_Authorize(t *testing.T) {
	var (
		err      error
		roleAuth = permission.NewAppserverPermissionAuthorizer(testutil.TestDbConn, db.NewQuerier(qx.New(testutil.TestDbConn)))
	)

	t.Run("ActionRead", func(t *testing.T) {
		t.Run(permission.SubActionListAppserverUserPermsission, func(t *testing.T) {
			t.Run("Successful:owner_can_list_permission_roles", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				sub := testutil.TestAppserverSub(t, nil, true)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: sub.AppserverID,
				})

				// ACT
				err = roleAuth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListAppserverUserPermsission)

				// ASSERT
				assert.Nil(t, err)
			})

			t.Run("Error:subscribed_user_cannot_list_permission_roles", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				user := testutil.TestAppuser(t, nil, false)
				server := testutil.TestAppserver(t, nil, false)
				testutil.TestAppserverSub(t, &qx.AppserverSub{AppuserID: user.ID, AppserverID: server.ID}, false)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: server.ID,
				})

				// ACT
				err = roleAuth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListAppserverUserPermsission)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, "(-5) Unauthorized", err.Error())
			})

			t.Run("Error:unsubscribed_user_cannot_list_permission_roles", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				appserver := testutil.TestAppserver(t, nil, false)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: appserver.ID,
				})

				// ACT
				err = roleAuth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListAppserverUserPermsission)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, "(-5) Unauthorized", err.Error())
			})

			t.Run("Error:db_error_on_owner_check", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				mockQuerier := new(testutil.MockQuerier)
				mockQuerier.On("GetAppserverById", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))
				mockRoleAuth := permission.NewAppserverPermissionAuthorizer(testutil.TestDbConn, mockQuerier)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: uuid.New(),
				})

				// ACT
				err = mockRoleAuth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListAppserverUserPermsission)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, "(-3) database error: db error", err.Error())
			})
		})
	})

	t.Run("ActionWrite", func(t *testing.T) {
		t.Run("Successful:owner_can_create_permission_role", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			userId, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
			user := testutil.TestAppuser(t, &qx.Appuser{ID: userId, Username: "foo"}, false)
			appserver := testutil.TestAppserver(t, &qx.Appserver{AppuserID: user.ID}, false)

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: appserver.ID,
			})

			// ACT
			err = roleAuth.Authorize(ctx, nil, permission.ActionWrite, permission.SubActionCreate)

			// ASSERT
			assert.Nil(t, err)
		})

		t.Run("Error:non_owner_cannot_create_permission_role", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			appserver := testutil.TestAppserver(t, nil, false)

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: appserver.ID,
			})

			// ACT
			err = roleAuth.Authorize(ctx, nil, permission.ActionWrite, permission.SubActionCreate)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})

		t.Run("Error:subscribed_user_cannot_create_permission_role", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			user := testutil.TestAppuser(t, nil, false)
			appserver := testutil.TestAppserver(t, nil, false)
			testutil.TestAppserverSub(t, &qx.AppserverSub{AppuserID: user.ID, AppserverID: appserver.ID}, false)

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: appserver.ID,
			})

			// ACT
			err = roleAuth.Authorize(ctx, nil, permission.ActionWrite, permission.SubActionCreate)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})

		t.Run("Error:db_error_on_owner_check_permission", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			mockQuerier := new(testutil.MockQuerier)
			mockQuerier.On("GetAppserverById", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))
			mockRoleAuth := permission.NewAppserverPermissionAuthorizer(testutil.TestDbConn, mockQuerier)

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: uuid.New(),
			})

			// ACT
			err = mockRoleAuth.Authorize(ctx, nil, permission.ActionWrite, permission.SubActionCreate)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-3) database error: db error", err.Error())
		})
	})

	t.Run("ActionDelete", func(t *testing.T) {
		t.Run("Successful:owner_can_delete_permission_role", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			testutil.TestAppserverSub(t, nil, true)
			role := testutil.TestAppserverPermission(t, nil, true)

			idStr := role.ID.String()

			// ACT
			err = roleAuth.Authorize(ctx, &idStr, permission.ActionDelete, permission.SubActionDelete)

			// ASSERT
			assert.Nil(t, err)
		})

		t.Run("Error:non_owner_cannot_delete_permission_role", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			role := testutil.TestAppserverPermission(t, nil, false)

			idStr := role.ID.String()

			// ACT
			err = roleAuth.Authorize(ctx, &idStr, permission.ActionDelete, permission.SubActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})

		t.Run("Error:subscribed_user_cannot_delete_permission_role", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			user := testutil.TestAppuser(t, nil, false)
			appserver := testutil.TestAppserver(t, nil, false)
			testutil.TestAppserverSub(t, &qx.AppserverSub{AppuserID: user.ID, AppserverID: appserver.ID}, false)
			role := testutil.TestAppserverPermission(t, &qx.AppserverPermission{AppserverID: appserver.ID, AppuserID: user.ID}, false)

			idStr := role.ID.String()

			// ACT
			err = roleAuth.Authorize(ctx, &idStr, permission.ActionDelete, permission.SubActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})

		t.Run("Error:db_error_on_owner_check", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			idStr := testutil.TestAppserverPermission(t, nil, true).ID.String()

			mockQuerier := new(testutil.MockQuerier)
			mockQuerier.On("GetAppserverPermissionById", mock.Anything, mock.Anything).Return(qx.AppserverPermission{}, nil)
			mockQuerier.On("GetAppserverById", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

			mockRoleAuth := permission.NewAppserverPermissionAuthorizer(testutil.TestDbConn, mockQuerier)

			// ACT
			err = mockRoleAuth.Authorize(ctx, &idStr, permission.ActionDelete, permission.SubActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-3) database error: db error", err.Error())
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
			err = roleAuth.Authorize(badCtx, nil, permission.ActionRead, permission.SubActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-1) invalid uuid", err.Error())
		})

		t.Run("Error:invalid_object_id_format", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			badId := "invalid"

			// ACT
			err = roleAuth.Authorize(ctx, &badId, permission.ActionDelete, permission.SubActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-1) invalid uuid", err.Error())
		})

		t.Run("Error:object_id_not_found", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			nonExistentId := uuid.NewString()

			// ACT
			err = roleAuth.Authorize(ctx, &nonExistentId, permission.ActionDelete, permission.SubActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-2) resource not found", err.Error())
		})

		t.Run("Error:undefined_permission_defaults_to_error", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			role := testutil.TestAppserverPermission(t, nil, false)
			idStr := role.ID.String()

			// ACT
			err = roleAuth.Authorize(ctx, &idStr, permission.ActionWrite, "random-unknown-action")

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})
	})
}
