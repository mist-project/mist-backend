package permission_test

import (
	"context"
	"os"
	"testing"

	"mist/src/middleware"
	"mist/src/permission"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
	"mist/src/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAppserverAuthorizer_Authorize(t *testing.T) {

	var (
		err        error
		authorizer = permission.NewAppserverAuthorizer(testutil.TestDbConn, db.NewQuerier(qx.New(testutil.TestDbConn)))
		ctx        = testutil.Setup(t, func() {})
	)

	t.Run("ActionRead", func(t *testing.T) {

		t.Run("Successful:does_not_error_on_detail", func(t *testing.T) {

			// ACT
			err = authorizer.Authorize(ctx, nil, permission.ActionRead, "detail")

			// ASSERT
			assert.Nil(t, err)
		})

		t.Run("Successful:does_not_error_on_other_read_actions", func(t *testing.T) {

			// ACT
			err = authorizer.Authorize(ctx, nil, permission.ActionRead, "n/a")

			// ASSERT
			assert.Nil(t, err)
		})
	})

	t.Run("ActionWrite", func(t *testing.T) {

		t.Run("Successful:does_not_error_on_create", func(t *testing.T) {

			// ACT
			err = authorizer.Authorize(ctx, nil, permission.ActionWrite, "create")

			// ASSERT
			assert.Nil(t, err)
		})
	})

	t.Run("ActionDelete", func(t *testing.T) {

		t.Run("Successful:owner_can_delete_server", func(t *testing.T) {

			// ARRANGE
			userID, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
			testutil.TestAppuser(t, &qx.Appuser{ID: userID, Username: "foo"})
			appserver := testutil.TestAppserver(t, &qx.Appserver{Name: "bar", AppuserID: userID})
			idStr := appserver.ID.String()

			// ACT
			err = authorizer.Authorize(ctx, &idStr, permission.ActionDelete, "delete")

			// ASSERT
			assert.Nil(t, err)
		})

		t.Run("Error:non_owner_cannot_delete_server", func(t *testing.T) {

			// ARRANGE
			appserver := testutil.TestAppserver(t, nil)
			idStr := appserver.ID.String()

			// ACT
			err = authorizer.Authorize(ctx, &idStr, permission.ActionDelete, "delete")

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})
	})

	t.Run("Errors", func(t *testing.T) {

		t.Run("Error:invalid_user_id_in_context", func(t *testing.T) {
			// ARRANGE
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
			err = authorizer.Authorize(badCtx, nil, permission.ActionDelete, "delete")

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-1) invalid uuid", err.Error())
		})

		t.Run("Error:invalid_object_id_format", func(t *testing.T) {

			// ARRANGE
			badId := "invalid"

			// ACT
			err = authorizer.Authorize(ctx, &badId, permission.ActionDelete, "delete")

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-1) invalid uuid", err.Error())
		})

		t.Run("Error:object_id_not_found", func(t *testing.T) {

			// ARRANGE
			nonExistentId := uuid.NewString()

			// ACT
			err = authorizer.Authorize(ctx, &nonExistentId, permission.ActionDelete, "delete")

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-2) resource not found", err.Error())
		})

		t.Run("Error:undefined_permission_defaults_to_error", func(t *testing.T) {

			// ARRANGE
			appserver := testutil.TestAppserver(t, nil)
			idStr := appserver.ID.String()

			// ACT
			err = authorizer.Authorize(ctx, &idStr, permission.ActionWrite, "update")

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})
	})
}
