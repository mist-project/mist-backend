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

func TestAppserverRoleSubAuthorizer_Authorize(t *testing.T) {
	var (
		err         error
		roleSubAuth = permission.NewAppserverRoleSubAuthorizer(testutil.TestDbConn, db.NewQuerier(qx.New(testutil.TestDbConn)))
	)

	t.Run("ActionRead", func(t *testing.T) {
		t.Run(permission.SubActionListAppserverUserRoleSubs, func(t *testing.T) {
			t.Run("Successful:subscribed_user_can_list_role_subs", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				sub := testutil.TestAppserverSub(t, nil, true)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: sub.AppserverID,
				})

				// ACT
				err = roleSubAuth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListAppserverUserRoleSubs)

				// ASSERT
				assert.Nil(t, err)
			})

			t.Run("Error:unsubscribed_user_cannot_list_role_subs", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				appserver := testutil.TestAppserver(t, nil, false)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: appserver.ID,
				})

				// ACT
				err = roleSubAuth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListAppserverUserRoleSubs)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, "(-5) Unauthorized", err.Error())
			})

			t.Run("Error:db_error_on_sub_check", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				mockQuerier := new(testutil.MockQuerier)
				mockQuerier.On("FilterAppserverSub", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))
				mockRoleSubAuth := permission.NewAppserverRoleSubAuthorizer(testutil.TestDbConn, mockQuerier)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: uuid.New(),
				})

				// ACT
				err = mockRoleSubAuth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListAppserverUserRoleSubs)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, "(-3) database error: db error", err.Error())
			})
		})
	})

	t.Run("ActionWrite", func(t *testing.T) {
		t.Run("Successful:owner_can_create_role_sub", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			userId, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
			user := testutil.TestAppuser(t, &qx.Appuser{ID: userId, Username: "foo"}, false)
			appserver := testutil.TestAppserver(t, &qx.Appserver{AppuserID: user.ID}, false)

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: appserver.ID,
			})

			// ACT
			err = roleSubAuth.Authorize(ctx, nil, permission.ActionWrite, permission.SubActionCreate)

			// ASSERT
			assert.Nil(t, err)
		})

		t.Run("Error:non_owner_cannot_create_role_sub", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			appserver := testutil.TestAppserver(t, nil, false)

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: appserver.ID,
			})

			// ACT
			err = roleSubAuth.Authorize(ctx, nil, permission.ActionWrite, permission.SubActionCreate)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})

		t.Run("Error:db_error_on_owner_check", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			mockQuerier := new(testutil.MockQuerier)
			mockQuerier.On("GetAppserverById", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))
			mockRoleSubAuth := permission.NewAppserverRoleSubAuthorizer(testutil.TestDbConn, mockQuerier)

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: uuid.New(),
			})

			// ACT
			err = mockRoleSubAuth.Authorize(ctx, nil, permission.ActionWrite, permission.SubActionCreate)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-3) database error: db error", err.Error())
		})
	})

	t.Run("ActionDelete", func(t *testing.T) {
		t.Run("Successful:owner_can_delete_role_sub", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			testutil.TestAppserverSub(t, nil, true)
			roleSub := testutil.TestAppserverRoleSub(t, nil, true)

			idStr := roleSub.ID.String()

			// ACT
			err = roleSubAuth.Authorize(ctx, &idStr, permission.ActionDelete, permission.SubActionDelete)

			// ASSERT
			assert.Nil(t, err)
		})

		t.Run("Error:non_owner_cannot_delete_role_sub", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			roleSub := testutil.TestAppserverRoleSub(t, nil, false)

			idStr := roleSub.ID.String()

			// ACT
			err = roleSubAuth.Authorize(ctx, &idStr, permission.ActionDelete, permission.SubActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})

		t.Run("Error:db_error_on_owner_check", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			roleSubId := uuid.NewString()

			mockQuerier := new(testutil.MockQuerier)
			mockQuerier.On("GetAppserverRoleSubById", mock.Anything, mock.Anything).Return(qx.AppserverRoleSub{}, nil)
			mockQuerier.On("GetAppserverById", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

			mockRoleSubAuth := permission.NewAppserverRoleSubAuthorizer(testutil.TestDbConn, mockQuerier)

			// ACT
			err = mockRoleSubAuth.Authorize(ctx, &roleSubId, permission.ActionDelete, permission.SubActionDelete)

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
			err = roleSubAuth.Authorize(badCtx, nil, permission.ActionRead, permission.SubActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-1) invalid uuid", err.Error())
		})

		t.Run("Error:invalid_object_id_format", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			badId := "invalid"

			// ACT
			err = roleSubAuth.Authorize(ctx, &badId, permission.ActionDelete, permission.SubActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-1) invalid uuid", err.Error())
		})

		t.Run("Error:object_id_not_found", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			nonExistentId := uuid.NewString()

			// ACT
			err = roleSubAuth.Authorize(ctx, &nonExistentId, permission.ActionDelete, permission.SubActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-2) resource not found", err.Error())
		})

		t.Run("Error:undefined_permission_defaults_to_error", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			roleSub := testutil.TestAppserverRoleSub(t, nil, false)
			idStr := roleSub.ID.String()

			// ACT
			err = roleSubAuth.Authorize(ctx, &idStr, permission.ActionWrite, "random-unknown-action")

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})
	})
}
