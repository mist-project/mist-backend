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

func TestQuerier_CreateAppserver(t *testing.T) {
	t.Run("Success:create_appserver", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		u := factory.NewFactory(ctx, db).Appuser(t, 0, nil)

		params := qx.CreateAppserverParams{
			Name:      "Test Server",
			AppuserID: u.ID,
		}

		// ACT
		server, err := db.CreateAppserver(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, params.Name, server.Name)
		assert.Equal(t, params.AppuserID, server.AppuserID)
	})
}

func TestQuerier_GetAppserverById(t *testing.T) {
	t.Run("Success:get_appserver_by_id", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		server := factory.NewFactory(ctx, db).Appserver(t, 0, nil)

		// ACT
		result, err := db.GetAppserverById(ctx, server.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, server.ID, result.ID)
	})

	t.Run("Error:appserver_does_not_exist", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		// ACT
		_, err := db.GetAppserverById(ctx, uuid.New())

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})
}

func TestQuerier_DeleteAppserver(t *testing.T) {
	t.Run("Success:delete_appserver", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		server := factory.NewFactory(ctx, db).Appserver(t, 0, nil)

		// ACT
		count, err := db.DeleteAppserver(ctx, server.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Error:appserver_does_not_exist", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		// ACT
		count, err := db.DeleteAppserver(ctx, uuid.New())

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("Success:deletes_all_relationships", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		f := factory.NewFactory(ctx, db)
		server := f.Appserver(t, 0, nil)
		sub := f.AppserverSub(t, 0, nil)
		role := f.AppserverRole(t, 0, nil)
		roleSub := f.AppserverRoleSub(t, 0, nil)
		channel := f.Channel(t, 0, nil)
		channelRole := f.ChannelRole(t, 0, nil)

		// ACT
		count, err := db.DeleteAppserver(ctx, server.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		q := db
		// Verify that the AppserverRoleSub is deleted
		_, err = q.GetAppserverRoleSubById(ctx, roleSub.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")

		// Verify that the ChannelRole is deleted
		_, err = q.GetChannelRoleById(ctx, channelRole.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")

		// Verify that the AppserverRole is deleted
		_, err = q.GetAppserverRoleById(ctx, role.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")

		// Verify that the Channel is deleted
		_, err = q.GetChannelById(ctx, channel.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")

		// Verify that the Appserver is deleted
		_, err = q.GetAppserverById(ctx, server.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")

		// Verify that the AppserverSub is deleted
		_, err = q.GetAppserverSubById(ctx, sub.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})
}

func TestQuerier_ListAppservers(t *testing.T) {
	t.Run("Success:list_appservers_by_user", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		server := factory.NewFactory(ctx, db).Appserver(t, 0, nil)

		params := qx.ListAppserversParams{
			AppuserID: server.AppuserID,
			Name:      pgtype.Text{Valid: false},
		}

		// ACT
		results, err := db.ListAppservers(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("Success:list_appservers_by_user_and_name", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		server := factory.NewFactory(ctx, db).Appserver(t, 0, nil)

		params := qx.ListAppserversParams{
			AppuserID: server.AppuserID,
			Name:      pgtype.Text{String: server.Name, Valid: true},
		}

		// ACT
		results, err := db.ListAppservers(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, server.ID, results[0].ID)
	})

	t.Run("Success:list_appservers_no_results", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		params := qx.ListAppserversParams{
			AppuserID: uuid.New(),
			Name:      pgtype.Text{Valid: false},
		}

		// ACT
		results, err := db.ListAppservers(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}
