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

func TestAppserverSubAuthorizer_Authorize(t *testing.T) {
	var (
		err     error
		subAuth = permission.NewAppserverSubAuthorizer(testutil.TestDbConn, db.NewQuerier(qx.New(testutil.TestDbConn)))
	)

	t.Run("ActionRead", func(t *testing.T) {
		t.Run(permission.SubActionListAppserverUserSubs, func(t *testing.T) {
			t.Run("Successful:subscribed_user_can_list", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				sub := testutil.TestAppserverSub(t, nil, true)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: sub.AppserverID,
				})

				// ACT
				err = subAuth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListAppserverUserSubs)

				// ASSERT
				assert.Nil(t, err)
			})

			t.Run("Error:unsubscribed_user_cannot_list", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				appserver := testutil.TestAppserver(t, nil, false)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: appserver.ID,
				})

				// ACT
				err = subAuth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListAppserverUserSubs)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, "(-5) Unauthorized", err.Error())
			})

			t.Run("Error:db_error_on_sub_check", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				mockQuerier := new(testutil.MockQuerier)
				mockQuerier.On("FilterAppserverSub", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))
				mockSubAuth := permission.NewAppserverSubAuthorizer(testutil.TestDbConn, mockQuerier)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: uuid.New(),
				})

				// ACT
				err = mockSubAuth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListAppserverUserSubs)

				// ASSERT
				assert.NotNil(t, err)
				assert.Equal(t, "(-3) database error: db error", err.Error())
			})
		})

		t.Run(permission.SubActionListUserServerSubs, func(t *testing.T) {
			t.Run("Successful:subscribed_user_can_list", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})
				sub := testutil.TestAppserverSub(t, nil, true)

				ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
					AppserverId: sub.AppserverID,
				})

				// ACT
				err = subAuth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListUserServerSubs)

				// ASSERT
				assert.Nil(t, err)
			})

			t.Run("Error:unsubscribed_user_can_list", func(t *testing.T) {
				// ARRANGE
				ctx := testutil.Setup(t, func() {})

				// ACT
				err = subAuth.Authorize(ctx, nil, permission.ActionRead, permission.SubActionListUserServerSubs)

				// ASSERT
				assert.Nil(t, err)
			})
		})
	})

	t.Run("ActionWrite", func(t *testing.T) {
		t.Run("Successful:anyone_can_create_appserver_sub", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			appserver := testutil.TestAppserver(t, nil, false)

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: appserver.ID,
			})

			// ACT
			err = subAuth.Authorize(ctx, nil, permission.ActionWrite, permission.SubActionCreate)

			// ASSERT
			assert.Nil(t, err)
		})
	})

	t.Run("ActionDelete", func(t *testing.T) {
		t.Run("Successful:owner_can_delete_another_user_sub", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			user := testutil.TestAppuser(t, nil, false)
			server := testutil.TestAppserver(t, nil, true)
			sub := testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: server.ID, AppuserID: user.ID}, false)

			idStr := sub.ID.String()

			// ACT
			err = subAuth.Authorize(ctx, &idStr, permission.ActionDelete, permission.SubActionDelete)

			// ASSERT
			assert.Nil(t, err)
		})

		t.Run("Successful:object_owner_can_delete_its_own_subscription", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			appuser := testutil.TestAppuser(t, nil, true)
			server := testutil.TestAppserver(t, nil, false)
			sub := testutil.TestAppserverSub(t, &qx.AppserverSub{AppuserID: appuser.ID, AppserverID: server.ID}, false)

			idStr := sub.ID.String()

			// ACT
			err = subAuth.Authorize(ctx, &idStr, permission.ActionDelete, permission.SubActionDelete)

			// ASSERT
			assert.Nil(t, err)
		})

		t.Run("Error:non_owner_cannot_delete_other_user_sub", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			sub := testutil.TestAppserverSub(t, nil, false)

			idStr := sub.ID.String()

			// ACT
			err = subAuth.Authorize(ctx, &idStr, permission.ActionDelete, permission.SubActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})

		t.Run("Error:db_error_on_owner_check", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			subId := uuid.NewString()

			mockQuerier := new(testutil.MockQuerier)
			mockQuerier.On("GetAppserverSubById", mock.Anything, mock.Anything).Return(qx.AppserverSub{}, nil)
			mockQuerier.On("GetAppserverById", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("db error"))

			mockSubAuth := permission.NewAppserverSubAuthorizer(testutil.TestDbConn, mockQuerier)

			// ACT
			err = mockSubAuth.Authorize(ctx, &subId, permission.ActionDelete, permission.SubActionDelete)

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
			err = subAuth.Authorize(badCtx, nil, permission.ActionRead, permission.SubActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-1) invalid uuid", err.Error())
		})

		t.Run("Error:invalid_object_id_format", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			badId := "invalid"

			// ACT
			err = subAuth.Authorize(ctx, &badId, permission.ActionDelete, permission.SubActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-1) invalid uuid", err.Error())
		})

		t.Run("Error:object_id_not_found", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			nonExistentId := uuid.NewString()

			// ACT
			err = subAuth.Authorize(ctx, &nonExistentId, permission.ActionDelete, permission.SubActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-2) resource not found", err.Error())
		})

		t.Run("Error:undefined_permission_defaults_to_error", func(t *testing.T) {
			// ARRANGE
			ctx := testutil.Setup(t, func() {})
			sub := testutil.TestAppserverSub(t, nil, false)
			idStr := sub.ID.String()

			// ACT
			err = subAuth.Authorize(ctx, &idStr, permission.ActionWrite, "random-action")

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, "(-5) Unauthorized", err.Error())
		})
	})
}
