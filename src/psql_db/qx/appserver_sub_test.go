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
	t.Run("Successful:create_appserver_sub", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		user := testutil.TestAppuser(t, nil, false)
		server := testutil.TestAppserver(t, nil, false)

		params := qx.CreateAppserverSubParams{
			AppserverID: server.ID,
			AppuserID:   user.ID,
		}

		// ACT
		sub, err := qx.New(testutil.TestDbConn).CreateAppserverSub(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, user.ID, sub.AppuserID)
		assert.Equal(t, server.ID, sub.AppserverID)
	})
}

func TestQuerier_DeleteAppserverSub(t *testing.T) {
	t.Run("Successful:delete_appserver_sub", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		sub := testutil.TestAppserverSub(t, nil, false)

		// ACT
		count, err := qx.New(testutil.TestDbConn).DeleteAppserverSub(ctx, sub.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Error:sub_does_not_exist", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		count, err := qx.New(testutil.TestDbConn).DeleteAppserverSub(ctx, uuid.New())
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("Successful:deleting_appserver_sub_deletes_appserver_role_subs", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		su := factory.UserAppserverSub(t)
		role := testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "foo", AppserverID: su.Server.ID}, false)
		_, err := qx.New(testutil.TestDbConn).CreateAppserverRoleSub(ctx, qx.CreateAppserverRoleSubParams{
			AppuserID:       su.User.ID,
			AppserverID:     su.Server.ID,
			AppserverRoleID: role.ID,
			AppserverSubID:  su.Sub.ID,
		})
		assert.NoError(t, err)

		// ACT
		count, err := qx.New(testutil.TestDbConn).DeleteAppserverSub(ctx, su.Sub.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		roleSubs, err := qx.New(testutil.TestDbConn).ListServerRoleSubs(ctx, su.Server.ID)
		assert.NoError(t, err)
		assert.Empty(t, roleSubs)
	})
}

func TestQuerier_FilterAppserverSub(t *testing.T) {
	t.Run("Successful:filter_appserver_sub", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		su := factory.UserAppserverSub(t)

		params := qx.FilterAppserverSubParams{
			AppuserID:   pgtype.UUID{Bytes: su.User.ID, Valid: true},
			AppserverID: pgtype.UUID{Bytes: su.Server.ID, Valid: true},
		}

		results, err := qx.New(testutil.TestDbConn).FilterAppserverSub(ctx, params)

		assert.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("Error:invalid_filter_params", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		params := qx.FilterAppserverSubParams{
			AppuserID:   pgtype.UUID{Valid: false},
			AppserverID: pgtype.UUID{Valid: false},
		}

		results, err := qx.New(testutil.TestDbConn).FilterAppserverSub(ctx, params)
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestQuerier_GetAppserverSubById(t *testing.T) {
	t.Run("Successful:get_appserver_sub_by_id", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		sub := testutil.TestAppserverSub(t, nil, false)

		result, err := qx.New(testutil.TestDbConn).GetAppserverSubById(ctx, sub.ID)

		assert.NoError(t, err)
		assert.Equal(t, sub.ID, result.ID)
	})

	t.Run("Error:sub_does_not_exist", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		_, err := qx.New(testutil.TestDbConn).GetAppserverSubById(ctx, uuid.New())
		assert.Error(t, err)
	})
}

func TestQuerier_ListAppserverUserSubs(t *testing.T) {
	t.Run("Successful:list_appserver_user_subs", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		su := factory.UserAppserverSub(t)
		sub2 := testutil.TestAppserverSub(t, nil, false)

		results, err := qx.New(testutil.TestDbConn).ListAppserverUserSubs(ctx, su.Server.ID)

		assert.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.Len(t, results, 1)
		assert.NotContains(t, results, sub2.ID)
	})
}

func TestQuerier_ListUserServerSubs(t *testing.T) {
	t.Run("Successful:list_user_server_subs", func(t *testing.T) {
		ctx := testutil.Setup(t, func() {})
		su := factory.UserAppserverSub(t)
		sub2 := testutil.TestAppserverSub(t, nil, false)

		results, err := qx.New(testutil.TestDbConn).ListUserServerSubs(ctx, su.User.ID)

		assert.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.Len(t, results, 1)
		assert.NotContains(t, results, sub2.ID)
	})
}
