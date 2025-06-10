package qx_test

import (
	"mist/src/psql_db/qx"
	"mist/src/testutil"
	"mist/src/testutil/factory"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestQuerier_ListServerRoleSubs(t *testing.T) {
	t.Run("Success:list_role_subs_by_appserver", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		f := factory.NewFactory(ctx, db)
		role1 := f.AppserverRoleSub(t, 0, nil)
		f.AppserverRoleSub(t, 0, nil)

		// ACT
		results, err := db.ListServerRoleSubs(ctx, role1.AppserverID)

		// ASSERT
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, role1.AppserverID, results[0].AppserverID)
	})
}

func TestQuerier_CreateAppserverRoleSub(t *testing.T) {
	t.Run("Success:create_appserver_role_sub", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverSub(t, ctx, db)
		role1 := factory.NewFactory(ctx, db).AppserverRole(t, 0, nil)

		// ACT
		r, err := db.CreateAppserverRoleSub(ctx, qx.CreateAppserverRoleSubParams{
			AppserverSubID:  su.Sub.ID,
			AppserverRoleID: role1.ID,
			AppuserID:       su.User.ID,
			AppserverID:     su.Server.ID,
		})

		// ASSERT
		assert.NoError(t, err)
		assert.NotNil(t, r)
	})

	t.Run("Error:when_appserver_sub_and_appserver_role_dont_belong_to_same_server_it_errors", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		f := factory.NewFactory(ctx, db)
		user := f.Appuser(t, 0, nil)
		sub := f.AppserverSub(t, 0, nil)
		role := f.AppserverRole(t, 1, nil)

		// ACT
		_, err := db.CreateAppserverRoleSub(ctx, qx.CreateAppserverRoleSubParams{
			AppserverSubID:  sub.ID,
			AppserverRoleID: role.ID,
			AppuserID:       user.ID,
			AppserverID:     role.AppserverID,
		})

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `violates foreign key constraint "appserver_role_sub_uk_server_and_sub"`)
	})
}

func TestQuerier_GetById(t *testing.T) {
	t.Run("Success:get_appserver_role_sub_by_id", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		role1 := factory.NewFactory(ctx, db).AppserverRoleSub(t, 0, nil)

		// ACT
		result, err := db.GetAppserverRoleSubById(ctx, role1.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, role1.ID, result.ID)
	})

	t.Run("Error:when_appserver_role_sub_does_not_exist_it_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		id := uuid.New()

		// ACT
		_, err := db.GetAppserverRoleSubById(ctx, id)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})

}

func TestQuerier_DeleteAppserverRoleSub(t *testing.T) {
	t.Run("Success:delete_appserver_role_sub", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		role1 := factory.NewFactory(ctx, db).AppserverRoleSub(t, 0, nil)

		// ACT
		count, err := db.DeleteAppserverRoleSub(ctx, role1.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Error:when_appserver_role_sub_does_not_exist_it_returns_zero_count", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		id := uuid.New()

		// ACT
		count, _ := db.DeleteAppserverRoleSub(ctx, id)

		// ASSERT
		assert.Equal(t, int64(0), count)
	})
}

func TestQuerier_FilterAppserverRoleSub(t *testing.T) {
	t.Run("Success:filter_by_all_fields", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		roleSub := factory.NewFactory(ctx, db).AppserverRoleSub(t, 0, nil)

		params := qx.FilterAppserverRoleSubParams{
			AppuserID:       pgtype.UUID{Bytes: roleSub.AppuserID, Valid: true},
			AppserverID:     pgtype.UUID{Bytes: roleSub.AppserverID, Valid: true},
			AppserverRoleID: pgtype.UUID{Bytes: roleSub.AppserverRoleID, Valid: true},
			AppserverSubID:  pgtype.UUID{Bytes: roleSub.AppserverSubID, Valid: true},
		}

		// ACT
		results, err := db.FilterAppserverRoleSub(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.Equal(t, roleSub.ID, results[0].ID)
	})

	t.Run("Success:filter_by_partial_fields", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		roleSub := factory.NewFactory(ctx, db).AppserverRoleSub(t, 0, nil)

		params := qx.FilterAppserverRoleSubParams{
			AppuserID:       pgtype.UUID{Bytes: roleSub.AppuserID, Valid: true},
			AppserverID:     pgtype.UUID{Valid: false},
			AppserverRoleID: pgtype.UUID{Valid: false},
			AppserverSubID:  pgtype.UUID{Valid: false},
		}

		// ACT
		results, err := db.FilterAppserverRoleSub(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("EmptyResult:when_filter_does_not_match", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		_ = factory.NewFactory(ctx, db).AppserverRoleSub(t, 0, nil)

		params := qx.FilterAppserverRoleSubParams{
			AppuserID:       pgtype.UUID{Bytes: uuid.New(), Valid: true},
			AppserverID:     pgtype.UUID{Valid: false},
			AppserverRoleID: pgtype.UUID{Valid: false},
			AppserverSubID:  pgtype.UUID{Valid: false},
		}

		// ACT
		results, err := db.FilterAppserverRoleSub(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}
