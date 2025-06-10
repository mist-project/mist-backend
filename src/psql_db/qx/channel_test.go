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

func TestQuerier_CreateChannel(t *testing.T) {
	t.Run("Success:create_channel", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		server := testutil.TestAppserver(t, nil, false)

		params := qx.CreateChannelParams{
			Name:        "general",
			AppserverID: server.ID,
			IsPrivate:   false,
		}

		// ACT
		ch, err := qx.New(testutil.TestDbConn).CreateChannel(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, params.Name, ch.Name)
		assert.Equal(t, params.AppserverID, ch.AppserverID)
	})
}

func TestQuerier_GetChannelById(t *testing.T) {
	t.Run("Success:get_channel_by_id", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		ch := testutil.TestChannel(t, nil, false)

		// ACT
		result, err := qx.New(testutil.TestDbConn).GetChannelById(ctx, ch.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, ch.ID, result.ID)
	})

	t.Run("Error:channel_does_not_exist", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		// ACT
		_, err := qx.New(testutil.TestDbConn).GetChannelById(ctx, uuid.New())

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})
}

func TestQuerier_DeleteChannel(t *testing.T) {
	t.Run("Success:delete_channel", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		ch := testutil.TestChannel(t, nil, false)

		// ACT
		count, err := qx.New(testutil.TestDbConn).DeleteChannel(ctx, ch.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Success:when_channel_is_deleted_it_removes_associated_channel_roles", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		ch := testutil.TestChannel(t, nil, false)
		role := testutil.TestAppserverRole(t, &qx.AppserverRole{AppserverID: ch.AppserverID}, false)
		channelRole := testutil.TestChannelRole(
			t, &qx.ChannelRole{
				ChannelID: ch.ID, AppserverRoleID: role.ID, AppserverID: ch.AppserverID,
			}, false)

		// ACT
		count, err := qx.New(testutil.TestDbConn).DeleteChannel(ctx, ch.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
		// Verify that the channel role is also deleted
		_, err = qx.New(testutil.TestDbConn).GetChannelRoleById(ctx, channelRole.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})

	t.Run("Error:channel_does_not_exist", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		// ACT
		count, err := qx.New(testutil.TestDbConn).DeleteChannel(ctx, uuid.New())

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestQuerier_FilterChannel(t *testing.T) {
	t.Run("Success:filter_channel", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		ch := testutil.TestChannel(t, nil, false)

		params := qx.FilterChannelParams{
			AppserverID: pgtype.UUID{Bytes: ch.AppserverID, Valid: true},
			IsPrivate:   pgtype.Bool{Bool: ch.IsPrivate, Valid: true},
		}

		// ACT
		results, err := qx.New(testutil.TestDbConn).FilterChannel(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("Error:filter_channel_no_results", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		testutil.TestChannel(t, nil, false)

		params := qx.FilterChannelParams{
			AppserverID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
			IsPrivate:   pgtype.Bool{Bool: false, Valid: true},
		}

		// ACT
		results, err := qx.New(testutil.TestDbConn).FilterChannel(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("Error:filter_channel_invalid_params", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		params := qx.FilterChannelParams{
			AppserverID: pgtype.UUID{Valid: false},
			IsPrivate:   pgtype.Bool{Valid: false},
		}

		// ACT
		results, err := qx.New(testutil.TestDbConn).FilterChannel(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestQuerier_GetChannelsIdIn(t *testing.T) {
	t.Run("Success:get_channels_id_in", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		ch := testutil.TestChannel(t, nil, false)
		ch2 := testutil.TestChannel(t, nil, false)

		// ACT
		results, err := qx.New(testutil.TestDbConn).GetChannelsIdIn(ctx, []uuid.UUID{ch.ID, ch2.ID})

		// ASSERT
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, ch.ID, results[0].ID)
		assert.Equal(t, ch2.ID, results[1].ID)
	})

	t.Run("Error:get_channels_id_in_no_results", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		// ACT
		results, err := qx.New(testutil.TestDbConn).GetChannelsIdIn(ctx, []uuid.UUID{uuid.New(), uuid.New()})

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestQuerier_ListServerChannels(t *testing.T) {
	t.Run("Success:list_server_channels", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		ch := testutil.TestChannel(t, nil, false)

		params := qx.ListServerChannelsParams{
			AppserverID: ch.AppserverID,
			Name:        pgtype.Text{String: ch.Name, Valid: true},
		}

		// ACT
		results, err := qx.New(testutil.TestDbConn).ListServerChannels(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("Error:list_server_channels_no_results", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		testutil.TestChannel(t, nil, false)

		params := qx.ListServerChannelsParams{
			AppserverID: uuid.New(),
			Name:        pgtype.Text{Valid: false},
		}

		// ACT
		results, err := qx.New(testutil.TestDbConn).ListServerChannels(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestQuerier_GetChannelsForUsers(t *testing.T) {
	t.Run("Success:get_channels_for_users", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverSub(t, ctx, db)
		user2 := testutil.TestAppuser(t, nil, false)
		channel1 := testutil.TestChannel(t, &qx.Channel{Name: "c1", AppserverID: su.Server.ID, IsPrivate: true}, false)
		_ = testutil.TestChannel(t, &qx.Channel{Name: "c2", AppserverID: su.Server.ID, IsPrivate: false}, false)
		role := testutil.TestAppserverRole(t, &qx.AppserverRole{AppserverID: su.Server.ID, Name: "test_role"}, false)
		testutil.TestChannelRole(
			t, &qx.ChannelRole{
				ChannelID: channel1.ID, AppserverRoleID: role.ID, AppserverID: su.Server.ID,
			}, false)

		testutil.TestAppserverRoleSub(
			t,
			&qx.AppserverRoleSub{
				AppuserID: su.User.ID, AppserverRoleID: role.ID, AppserverSubID: su.Sub.ID, AppserverID: su.Server.ID,
			}, false,
		)

		// ACT
		results, err := qx.New(testutil.TestDbConn).GetChannelsForUsers(ctx, qx.GetChannelsForUsersParams{
			Column1:     []uuid.UUID{su.User.ID, user2.ID},
			AppserverID: su.Server.ID,
		})

		// ASSERT
		users := make([]uuid.UUID, 0)
		userChannels := make(map[uuid.UUID][]uuid.UUID)
		assert.NoError(t, err)
		assert.NotEmpty(t, results)

		for _, r := range results {
			users = append(users, r.AppuserID)
			userChannels[r.AppuserID] = append(userChannels[r.AppuserID], r.ChannelID.Bytes)
		}
		assert.Contains(t, users, su.User.ID)
		assert.Contains(t, users, user2.ID)
		assert.Equal(t, 2, len(userChannels[su.User.ID]))
		assert.Equal(t, 1, len(userChannels[user2.ID]))
	})
}
