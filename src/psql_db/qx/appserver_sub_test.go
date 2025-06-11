package qx_test

import (
	"testing"

	"mist/src/psql_db/qx"
	"mist/src/testutil"
	"mist/src/testutil/factory"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestQuerier_CreateAppserverSub(t *testing.T) {
	t.Run("Success:create_appserver_sub", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		f := factory.NewFactory(ctx, db)
		user := f.Appuser(t, 0, nil)
		server := f.Appserver(t, 0, nil)

		params := qx.CreateAppserverSubParams{
			AppserverID: server.ID,
			AppuserID:   user.ID,
		}

		// ACT
		sub, err := db.CreateAppserverSub(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, user.ID, sub.AppuserID)
		assert.Equal(t, server.ID, sub.AppserverID)
	})
}

func TestQuerier_DeleteAppserverSub(t *testing.T) {
	t.Run("Success:delete_appserver_sub", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		sub := factory.NewFactory(ctx, db).AppserverSub(t, 0, nil)
		// ACT
		count, err := db.DeleteAppserverSub(ctx, sub.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Error:sub_does_not_exist", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		// ACT
		count, err := db.DeleteAppserverSub(ctx, uuid.New())

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("Success:deleting_appserver_sub_deletes_appserver_role_subs", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		f := factory.NewFactory(ctx, db)
		sub := f.AppserverSub(t, 0, nil)
		f.AppserverRole(t, 0, nil)
		f.AppserverRoleSub(t, 0, nil)

		// ACT
		count, err := db.DeleteAppserverSub(ctx, sub.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		roleSubs, err := db.ListServerRoleSubs(ctx, sub.AppserverID)
		assert.NoError(t, err)
		assert.Empty(t, roleSubs)
	})
}

func TestQuerier_FilterAppserverSub(t *testing.T) {
	t.Run("Success:filter_appserver_sub", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverSub(t, ctx, db)

		params := qx.FilterAppserverSubParams{
			AppuserID:   pgtype.UUID{Bytes: su.User.ID, Valid: true},
			AppserverID: pgtype.UUID{Bytes: su.Server.ID, Valid: true},
		}

		// ACT
		results, err := db.FilterAppserverSub(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("Error:invalid_filter_params", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		params := qx.FilterAppserverSubParams{
			AppuserID:   pgtype.UUID{Valid: false},
			AppserverID: pgtype.UUID{Valid: false},
		}

		// ACT
		results, err := db.FilterAppserverSub(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestQuerier_GetAppserverSubById(t *testing.T) {
	t.Run("Success:get_appserver_sub_by_id", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		sub := factory.NewFactory(ctx, db).AppserverSub(t, 0, nil)

		// ACT
		result, err := db.GetAppserverSubById(ctx, sub.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, sub.ID, result.ID)
	})

	t.Run("Error:sub_does_not_exist", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		// ACT
		_, err := db.GetAppserverSubById(ctx, uuid.New())

		// ASSERT
		assert.Error(t, err)
	})
}

func TestQuerier_ListAppserverUserSubs(t *testing.T) {
	t.Run("Success:list_appserver_user_subs", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverSub(t, ctx, db)
		sub2 := factory.NewFactory(ctx, db).AppserverSub(t, 1, nil)

		// ACT
		results, err := db.ListAppserverUserSubs(ctx, su.Server.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.Len(t, results, 1)
		assert.NotContains(t, results, sub2.ID)
	})
}

func TestQuerier_ListUserServerSubs(t *testing.T) {
	t.Run("Success:list_user_server_subs", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverSub(t, ctx, db)
		sub2 := factory.NewFactory(ctx, db).AppserverSub(t, 1, nil)

		// ACT
		results, err := db.ListUserServerSubs(ctx, su.User.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.Len(t, results, 1)
		assert.NotContains(t, results, sub2.ID)
	})
}
