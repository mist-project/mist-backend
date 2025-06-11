package permission_test

import (
	"context"
	"os"
	"testing"

	"mist/src/faults"
	"mist/src/middleware"
	"mist/src/permission"
	"mist/src/testutil"
	"mist/src/testutil/factory"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAppserverAuthorizer_Authorize(t *testing.T) {

	var (
		err error
	)

	t.Run("ActionRead", func(t *testing.T) {
		t.Run(string(permission.ActionRead), func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})

			t.Run("Success:unsubscribe_user_has_access", func(t *testing.T) {
				// ACT
				err = permission.NewAppserverAuthorizer(db).Authorize(ctx, nil, permission.ActionRead)

				// ASSERT
				assert.Nil(t, err)
			})
		})
	})

	t.Run("ActionCreate", func(t *testing.T) {

		t.Run("Success:any_user_can_create_appserver", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})

			// ACT
			err = permission.NewAppserverAuthorizer(db).Authorize(ctx, nil, permission.ActionCreate)

			// ASSERT
			assert.Nil(t, err)
		})
	})

	t.Run("ActionDelete", func(t *testing.T) {

		t.Run("Success:owner_can_delete_server", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			su := factory.UserAppserverOwner(t, ctx, db)
			idStr := su.Server.ID.String()

			// ACT
			err = permission.NewAppserverAuthorizer(db).Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.Nil(t, err)
		})

		t.Run("Error:user_with_manage_appserver_permission_cannot_delete_server", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			tu := factory.UserAppserverWithAllPermissions(t, ctx, db)
			idStr := tu.Server.ID.String()

			// ACT
			err = permission.NewAppserverAuthorizer(db).Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "user is not allowed to manage server")
		})

		t.Run("Error:user_without_manage_appserver_permission_cannot_delete_server", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			tu := factory.UserAppserverSub(t, ctx, db)
			idStr := tu.Server.ID.String()

			// ACT
			err = permission.NewAppserverAuthorizer(db).Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "user is not allowed to manage server")
		})

		t.Run("Error:non_owner_cannot_delete_server", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			su := factory.UserAppserverSub(t, ctx, db)
			idStr := su.Server.ID.String()

			// ACT
			err = permission.NewAppserverAuthorizer(db).Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "user is not allowed to manage server")
		})
	})

	t.Run("Errors", func(t *testing.T) {
		t.Run("Error:invalid_user_id_in_context", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			_, claims := testutil.CreateJwtToken(
				t,
				&testutil.CreateTokenParams{
					Iss:       os.Getenv("MIST_API_JWT_ISSUER"),
					Aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
					SecretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
					UserId:    "invalid",
				},
			)
			badCtx := context.WithValue(ctx, middleware.JwtClaimsK, claims)

			// ACT
			err = permission.NewAppserverAuthorizer(db).Authorize(badCtx, nil, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "invalid user id: invalid")
		})

		t.Run("Error:invalid_object_id_format", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			badId := "invalid"

			// ACT
			err = permission.NewAppserverAuthorizer(db).Authorize(ctx, &badId, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.ValidationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "invalid uuid pars")
		})

		t.Run("Error:object_id_not_found", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			nonExistentId := uuid.NewString()

			// ACT
			err = permission.NewAppserverAuthorizer(db).Authorize(ctx, &nonExistentId, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
		})

		t.Run("Error:nil_object_errors", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			var nilObj *string

			// ACT
			err = permission.NewAppserverAuthorizer(db).Authorize(ctx, nilObj, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "object id is required for action: delete")
		})
	})
}
