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
	"mist/src/psql_db/qx"
	"mist/src/testutil"
	"mist/src/testutil/factory"
)

func TestAppserverSubAuthorizer_Authorize(t *testing.T) {
	var (
		err error
	)

	t.Run("ActionRead", func(t *testing.T) {

		t.Run("Success:owner_can_read", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			tu := factory.UserAppserverOwner(t, ctx, db)
			idString := tu.Sub.ID.String()

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})

			// ACT
			err = permission.NewAppserverSubAuthorizer(db).Authorize(ctx, &idString, permission.ActionRead)

			// ASSERT
			assert.Nil(t, err)
		})

		t.Run("Success:subscribed_user_can_read", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			tu := factory.UserAppserverSub(t, ctx, db)
			idString := tu.Sub.ID.String()

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})

			// ACT
			err = permission.NewAppserverSubAuthorizer(db).Authorize(ctx, &idString, permission.ActionRead)

			// ASSERT
			assert.Nil(t, err)
		})

		t.Run("Error:unsubscribed_user_cannot_read", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			tu := factory.UserAppserverUnsub(t, ctx, db)
			idString := tu.Sub.ID.String()
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})

			// ACT
			err = permission.NewAppserverSubAuthorizer(db).Authorize(ctx, &idString, permission.ActionRead)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "user does not have permission to manage subscriptions")
		})
	})

	t.Run("ActionWrite", func(t *testing.T) {

		t.Run("Success:anyone_can_create_appserver_sub", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			tu := factory.UserAppserverUnsub(t, ctx, db)

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})

			// ACT
			err = permission.NewAppserverSubAuthorizer(db).Authorize(ctx, nil, permission.ActionCreate)

			// ASSERT
			assert.Nil(t, err)
		})
	})

	t.Run("ActionDelete", func(t *testing.T) {
		t.Run("Success:owner_can_delete_another_user_sub", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			tu := factory.UserAppserverOwner(t, ctx, db)
			f := factory.NewFactory(ctx, db)
			user := f.Appuser(t, 1, nil)
			sub := f.AppserverSub(t, 1, &qx.AppserverSub{
				AppserverID: tu.Server.ID,
				AppuserID:   user.ID,
			})
			idStr := sub.ID.String()
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})

			// ACT
			err = permission.NewAppserverSubAuthorizer(db).Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.Nil(t, err)
		})

		t.Run("Error:nobody_can_delete_owner_sub", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			tu := factory.UserAppserverOwner(t, ctx, db)

			idStr := tu.Sub.ID.String()

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})

			// ACT
			err = permission.NewAppserverSubAuthorizer(db).Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "cannot delete the owner's sub")
		})

		t.Run("Error:object_owner_can_delete_its_own_subscription", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			tu := factory.UserAppserverSub(t, ctx, db)

			idStr := tu.Sub.ID.String()

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})

			// ACT
			err = permission.NewAppserverSubAuthorizer(db).Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.Nil(t, err)
		})

		t.Run("Success:user_with_delete_permission_can_delete_sub", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			tu := factory.UserAppserverWithAllPermissions(t, ctx, db)
			f := factory.NewFactory(ctx, db)
			user := f.Appuser(t, 2, nil)
			sub := f.AppserverSub(t, 2, &qx.AppserverSub{
				AppserverID: tu.Server.ID,
				AppuserID:   user.ID,
			})

			idStr := sub.ID.String()

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})

			// ACT
			err = permission.NewAppserverSubAuthorizer(db).Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.Nil(t, err)
		})

		t.Run("Error:subscribed_user_without_permission_cannot_delete_other_user_sub", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			tu := factory.UserAppserverSub(t, ctx, db)
			f := factory.NewFactory(ctx, db)
			user := f.Appuser(t, 2, nil)
			sub := f.AppserverSub(t, 2, &qx.AppserverSub{
				AppserverID: tu.Server.ID,
				AppuserID:   user.ID,
			})
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: sub.AppserverID,
			})
			idStr := sub.ID.String()

			// ACT
			err = permission.NewAppserverSubAuthorizer(db).Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "user does not have permission to manage subscriptions")
		})
	})

	t.Run("Errors", func(t *testing.T) {
		t.Run("Error:invalid_userid_in_context", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			_, claims := testutil.CreateJwtToken(t, &testutil.CreateTokenParams{
				Iss:       os.Getenv("MIST_API_JWT_ISSUER"),
				Aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
				SecretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
				UserId:    "invalid",
			})
			badCtx := context.WithValue(ctx, middleware.JwtClaimsK, claims)

			// ACT
			err = permission.NewAppserverSubAuthorizer(db).Authorize(badCtx, nil, permission.ActionRead)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "invalid user id: invalid")
		})

		t.Run("Error:db_error_on_sub_check", func(t *testing.T) {
			// ARRANGE
			ctx, _ := testutil.Setup(t, func() {})
			serverId := uuid.New()
			mockQuerier := new(testutil.MockQuerier)
			mockQuerier.On("FilterAppserverSub", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("boom"))
			mockSubAuth := permission.NewAppserverSubAuthorizer(mockQuerier)

			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: serverId,
			})
			idStr := uuid.New().String()

			// ACT
			err = mockSubAuth.Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "failed to check user subscription")
			mockQuerier.AssertExpectations(t)
		})

		t.Run("Error:db_error_on_server_search", func(t *testing.T) {
			// ARRANGE
			ctx, _ := testutil.Setup(t, func() {})
			userId := uuid.New()
			serverId := uuid.New()
			subId := uuid.New()
			idStr := uuid.New().String()
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: serverId,
			})

			mockQuerier := new(testutil.MockQuerier)
			mockQuerier.On("FilterAppserverSub", mock.Anything, mock.Anything).Return([]qx.FilterAppserverSubRow{
				{
					ID:          subId,
					AppserverID: serverId,
					AppuserID:   userId,
				},
			}, nil)
			mockQuerier.On("GetAppserverSubById", mock.Anything, mock.Anything).Return(qx.AppserverSub{
				ID:          subId,
				AppserverID: serverId,
				AppuserID:   userId,
			}, nil)
			mockQuerier.On("GetAppserverById", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("boom"))

			mockSubAuth := permission.NewAppserverSubAuthorizer(mockQuerier)

			// ACT
			err = mockSubAuth.Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "failed to get appserver")
			mockQuerier.AssertExpectations(t)
		})

		t.Run("Error:db_error_on_user_permission_mask", func(t *testing.T) {
			// ARRANGE
			ctx, _ := testutil.Setup(t, func() {})
			userId := uuid.New()
			serverId := uuid.New()
			subId := uuid.New()
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: serverId,
			})
			idStr := uuid.New().String()

			mockQuerier := new(testutil.MockQuerier)
			mockQuerier.On("FilterAppserverSub", mock.Anything, mock.Anything).Return([]qx.FilterAppserverSubRow{
				{
					ID:          subId,
					AppserverID: serverId,
					AppuserID:   userId,
				},
			}, nil)
			mockQuerier.On("GetAppserverSubById", mock.Anything, mock.Anything).Return(qx.AppserverSub{
				ID:          subId,
				AppserverID: serverId,
				AppuserID:   userId,
			}, nil)
			mockQuerier.On("GetAppserverById", mock.Anything, mock.Anything).Return(qx.Appserver{
				ID:        serverId,
				AppuserID: uuid.New(),
			}, nil)
			mockQuerier.On("GetAppuserRoles", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("boom"))

			mockSubAuth := permission.NewAppserverSubAuthorizer(mockQuerier)

			// ACT
			err = mockSubAuth.Authorize(ctx, &idStr, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "failed to get user permissions")
			mockQuerier.AssertExpectations(t)
		})

		t.Run("Error:invalid_object_id_format", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			tu := factory.UserAppserverSub(t, ctx, db)
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})
			badId := "invalid"

			// ACT
			err = permission.NewAppserverSubAuthorizer(db).Authorize(ctx, &badId, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.ValidationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "invalid uuid")
		})

		t.Run("Error:invalid_server_id_format", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			factory.UserAppserverSub(t, ctx, db)
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, "invalid")

			// ACT
			err = permission.NewAppserverSubAuthorizer(db).Authorize(ctx, nil, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.AuthorizationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "invalid permission-context in context")
		})

		t.Run("Error:object_id_not_found", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			tu := factory.UserAppserverSub(t, ctx, db)
			nonExistentId := uuid.NewString()
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})

			// ACT
			err = permission.NewAppserverSubAuthorizer(db).Authorize(ctx, &nonExistentId, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.NotFoundMessage)
			testutil.AssertCustomErrorContains(t, err, "resource not found")
		})

		t.Run("Error:nil_object_errors", func(t *testing.T) {
			// ARRANGE
			ctx, db := testutil.Setup(t, func() {})
			tu := factory.UserAppserverSub(t, ctx, db)
			ctx = context.WithValue(ctx, permission.PermissionCtxKey, &permission.AppserverIdAuthCtx{
				AppserverId: tu.Server.ID,
			})
			var nilObj *string

			// ACT
			err = permission.NewAppserverSubAuthorizer(db).Authorize(ctx, nilObj, permission.ActionDelete)

			// ASSERT
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), faults.ValidationErrorMessage)
			testutil.AssertCustomErrorContains(t, err, "object id is nil")
		})
	})
}
