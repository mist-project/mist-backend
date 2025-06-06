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
	t.Run("Successful:create_appserver", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		user := testutil.TestAppuser(t, nil, false)

		params := qx.CreateAppserverParams{
			Name:      "Test Server",
			AppuserID: user.ID,
		}

		// ACT
		server, err := qx.New(testutil.TestDbConn).CreateAppserver(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, params.Name, server.Name)
		assert.Equal(t, params.AppuserID, server.AppuserID)
	})
}

func TestQuerier_GetAppserverById(t *testing.T) {
	t.Run("Successful:get_appserver_by_id", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		server := testutil.TestAppserver(t, nil, false)

		// ACT
		result, err := qx.New(testutil.TestDbConn).GetAppserverById(ctx, server.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, server.ID, result.ID)
	})

	t.Run("Error:appserver_does_not_exist", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		_, err := qx.New(testutil.TestDbConn).GetAppserverById(ctx, uuid.New())

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})
}

func TestQuerier_DeleteAppserver(t *testing.T) {
	t.Run("Successful:delete_appserver", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		server := testutil.TestAppserver(t, nil, false)

		// ACT
		count, err := qx.New(testutil.TestDbConn).DeleteAppserver(ctx, server.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Error:appserver_does_not_exist", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		count, err := qx.New(testutil.TestDbConn).DeleteAppserver(ctx, uuid.New())

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("Successful:deletes_all_relationships", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		tu := factory.UserAppserverSub(t)
		role := testutil.TestAppserverRole(t, &qx.AppserverRole{AppserverID: tu.Server.ID, Name: "foo"}, false)
		roleSub := testutil.TestAppserverRoleSub(t, &qx.AppserverRoleSub{
			AppuserID:       tu.User.ID,
			AppserverID:     tu.Server.ID,
			AppserverRoleID: role.ID,
			AppserverSubID:  tu.Sub.ID,
		}, false)

		channel := testutil.TestChannel(t, &qx.Channel{
			Name:        "general",
			AppserverID: tu.Server.ID,
			IsPrivate:   false,
		}, false)

		channelRole := testutil.TestChannelRole(t, &qx.ChannelRole{
			ChannelID:       channel.ID,
			AppserverRoleID: role.ID,
			AppserverID:     tu.Server.ID,
		}, false)

		// ACT
		count, err := qx.New(testutil.TestDbConn).DeleteAppserver(ctx, tu.Server.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		q := qx.New(testutil.TestDbConn)
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
		_, err = q.GetAppserverById(ctx, tu.Server.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")

		// Verify that the AppserverSub is deleted
		_, err = q.GetAppserverSubById(ctx, tu.Sub.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})
}

func TestQuerier_ListAppservers(t *testing.T) {
	t.Run("Successful:list_appservers_by_user", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		server := testutil.TestAppserver(t, nil, false)

		params := qx.ListAppserversParams{
			AppuserID: server.AppuserID,
			Name:      pgtype.Text{Valid: false},
		}

		// ACT
		results, err := qx.New(testutil.TestDbConn).ListAppservers(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("Successful:list_appservers_by_user_and_name", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		server := testutil.TestAppserver(t, nil, false)

		params := qx.ListAppserversParams{
			AppuserID: server.AppuserID,
			Name:      pgtype.Text{String: server.Name, Valid: true},
		}

		// ACT
		results, err := qx.New(testutil.TestDbConn).ListAppservers(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, server.ID, results[0].ID)
	})

	t.Run("Successful:list_appservers_no_results", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		params := qx.ListAppserversParams{
			AppuserID: uuid.New(),
			Name:      pgtype.Text{Valid: false},
		}

		// ACT
		results, err := qx.New(testutil.TestDbConn).ListAppservers(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}
