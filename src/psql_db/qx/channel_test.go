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
		ctx, db := testutil.Setup(t, func() {})
		server := factory.NewFactory(ctx, db).Appserver(t, 0, nil)

		params := qx.CreateChannelParams{
			Name:        "general",
			AppserverID: server.ID,
			IsPrivate:   false,
		}

		// ACT
		ch, err := db.CreateChannel(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, params.Name, ch.Name)
		assert.Equal(t, params.AppserverID, ch.AppserverID)
	})
}

func TestQuerier_GetChannelById(t *testing.T) {
	t.Run("Success:get_channel_by_id", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		ch := factory.NewFactory(ctx, db).Channel(t, 0, nil)

		// ACT
		result, err := db.GetChannelById(ctx, ch.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, ch.ID, result.ID)
	})

	t.Run("Error:channel_does_not_exist", func(t *testing.T) {
		// ARRANGE

		ctx, db := testutil.Setup(t, func() {})

		// ACT
		_, err := db.GetChannelById(ctx, uuid.New())

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})
}

func TestQuerier_DeleteChannel(t *testing.T) {
	t.Run("Success:delete_channel", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		ch := factory.NewFactory(ctx, db).Channel(t, 0, nil)

		// ACT
		count, err := db.DeleteChannel(ctx, ch.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Success:when_channel_is_deleted_it_removes_associated_channel_roles", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		f := factory.NewFactory(ctx, db)
		ch := f.Channel(t, 0, nil)
		f.AppserverRole(t, 0, nil)
		channelRole := f.ChannelRole(t, 0, nil)

		// ACT
		count, err := db.DeleteChannel(ctx, ch.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
		// Verify that the channel role is also deleted
		_, err = db.GetChannelRoleById(ctx, channelRole.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})

	t.Run("Error:channel_does_not_exist", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		// ACT
		count, err := db.DeleteChannel(ctx, uuid.New())

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestQuerier_FilterChannel(t *testing.T) {
	t.Run("Success:filter_channel", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		ch := factory.NewFactory(ctx, db).Channel(t, 0, nil)

		params := qx.FilterChannelParams{
			AppserverID: pgtype.UUID{Bytes: ch.AppserverID, Valid: true},
			IsPrivate:   pgtype.Bool{Bool: ch.IsPrivate, Valid: true},
		}

		// ACT
		results, err := db.FilterChannel(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("Error:filter_channel_no_results", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		factory.NewFactory(ctx, db).Channel(t, 0, nil)

		params := qx.FilterChannelParams{
			AppserverID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
			IsPrivate:   pgtype.Bool{Bool: false, Valid: true},
		}

		// ACT
		results, err := db.FilterChannel(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("Error:filter_channel_invalid_params", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		params := qx.FilterChannelParams{
			AppserverID: pgtype.UUID{Valid: false},
			IsPrivate:   pgtype.Bool{Valid: false},
		}

		// ACT
		results, err := db.FilterChannel(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestQuerier_GetChannelsIdIn(t *testing.T) {
	t.Run("Success:get_channels_id_in", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		f := factory.NewFactory(ctx, db)
		ch1 := f.Channel(t, 0, nil)
		ch2 := f.Channel(t, 1, nil)
		f.Channel(t, 2, nil)

		// ACT
		results, err := db.GetChannelsIdIn(ctx, []uuid.UUID{ch1.ID, ch2.ID})

		// ASSERT
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, ch1.ID, results[0].ID)
		assert.Equal(t, ch2.ID, results[1].ID)
	})

	t.Run("Error:get_channels_id_in_no_results", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		// ACT
		results, err := db.GetChannelsIdIn(ctx, []uuid.UUID{uuid.New(), uuid.New()})

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestQuerier_ListServerChannels(t *testing.T) {
	t.Run("Success:list_server_channels", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		ch := factory.NewFactory(ctx, db).Channel(t, 0, nil)

		params := qx.ListServerChannelsParams{
			AppserverID: ch.AppserverID,
			Name:        pgtype.Text{String: ch.Name, Valid: true},
		}

		// ACT
		results, err := db.ListServerChannels(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("Error:list_server_channels_no_results", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		f := factory.NewFactory(ctx, db)
		f.Channel(t, 0, nil)

		params := qx.ListServerChannelsParams{
			AppserverID: uuid.New(),
			Name:        pgtype.Text{Valid: false},
		}

		// ACT
		results, err := db.ListServerChannels(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestQuerier_GetChannelsForUsers(t *testing.T) {
	t.Run("Success:get_channels_for_users", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		f := factory.NewFactory(ctx, db)
		server := f.Appserver(t, 0, nil)
		f.Channel(t, 0, &qx.Channel{Name: "c1", AppserverID: server.ID, IsPrivate: true})
		f.Channel(t, 1, &qx.Channel{Name: "c2", AppserverID: server.ID, IsPrivate: false})
		f.AppserverRole(t, 0, nil)
		user1 := f.Appuser(t, 0, nil)
		user2 := f.Appuser(t, 1, nil)
		f.AppserverSub(t, 0, nil)
		f.AppserverSub(t, 1, &qx.AppserverSub{AppuserID: user2.ID, AppserverID: server.ID})
		f.AppserverRoleSub(t, 0, nil)
		f.ChannelRole(t, 0, nil)

		// ACT
		results, err := db.GetChannelsForUsers(ctx, qx.GetChannelsForUsersParams{
			Column1:     []uuid.UUID{user1.ID, user2.ID},
			AppserverID: server.ID,
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
		assert.Contains(t, users, user1.ID)
		assert.Contains(t, users, user2.ID)
		assert.Equal(t, 2, len(userChannels[user1.ID]))
		assert.Equal(t, 1, len(userChannels[user2.ID]))
	})
}
