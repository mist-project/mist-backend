package permission_test

import (
	"context"
	"mist/src/middleware"
	"mist/src/permission"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
	"mist/src/testutil"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAppserverAuthorizer_Authorize(t *testing.T) {

	var err error
	authorizer := permission.NewAppserverAuthorizer(testutil.TestDbConn, db.NewQuerier(qx.New(testutil.TestDbConn)))
	ctx := testutil.Setup(t, func() {})

	t.Run("Successful:read_action_does_not_error", func(t *testing.T) {
		// ACT
		err = authorizer.Authorize(ctx, nil, permission.ActionRead, "n/a")

		// ASSERT
		assert.Nil(t, err)
	})

	t.Run("Successful:create_subaction_does_not_error", func(t *testing.T) {
		// ACT
		err = authorizer.Authorize(ctx, nil, permission.ActionWrite, "create")

		// ASSERT
		assert.Nil(t, err)
	})

	t.Run("Successful:owner_of_server_can_delete", func(t *testing.T) {
		// ARRANGE
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		testutil.TestAppuser(t, &qx.Appuser{ID: parsedUid, Username: "foo"})
		appserver := testutil.TestAppserver(t, &qx.Appserver{Name: "bar", AppuserID: parsedUid})
		idStr := appserver.ID.String()

		// ACT
		err = authorizer.Authorize(ctx, &idStr, permission.ActionDelete, "delete")

		// ASSERT
		assert.Nil(t, err)
	})

	t.Run("Error:on_delete_only_owners_can_delete_server", func(t *testing.T) {
		// ARRANGE
		appserver := testutil.TestAppserver(t, nil)
		idStr := appserver.ID.String()

		// ACT
		err = authorizer.Authorize(ctx, &idStr, permission.ActionDelete, "delete")

		// ASSERT
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "(-5) Unauthorized")
	})

	t.Run("Error:invalid_userid_in_context_errors", func(t *testing.T) {
		// ACT
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

		err = authorizer.Authorize(badCtx, nil, permission.ActionDelete, "delete")

		// ASSERT
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "(-1) invalid uuid")
	})

	t.Run("Error:invalid_uuid_when_object_id_present_errors", func(t *testing.T) {
		// ARRANGE
		idStr := "invalid"

		// ACT
		err = authorizer.Authorize(ctx, &idStr, permission.ActionDelete, "delete")

		// ASSERT
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "(-1) invalid uuid")
	})
	t.Run("Error:when_object_id_is_provided_but_object_not_found_it_errors", func(t *testing.T) {
		idStr := uuid.NewString()

		// ACT
		err = authorizer.Authorize(ctx, &idStr, permission.ActionDelete, "delete")

		// ASSERT
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "(-2) resource not found")
	})

	t.Run("Error:undefined_permission_defaults_to_error", func(t *testing.T) {
		appserver := testutil.TestAppserver(t, nil)
		idStr := appserver.ID.String()

		// ACT
		err = authorizer.Authorize(ctx, &idStr, permission.ActionWrite, "update")

		// ASSERT
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "(-5) Unauthorized")
	})
}
