package qx_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"mist/src/psql_db/qx"
	"mist/src/testutil"
)

func TestQuerier_CreateChannelRole(t *testing.T) {
	t.Run("Success:create_channel_role", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		ch := testutil.TestChannel(t, nil, false)
		role := testutil.TestAppserverRole(t, &qx.AppserverRole{AppserverID: ch.AppserverID}, false)

		params := qx.CreateChannelRoleParams{
			ChannelID:       ch.ID,
			AppserverRoleID: role.ID,
			AppserverID:     ch.AppserverID,
		}

		// ACT
		cr, err := qx.New(testutil.TestDbConn).CreateChannelRole(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, ch.ID, cr.ChannelID)
		assert.Equal(t, role.ID, cr.AppserverRoleID)
		assert.Equal(t, ch.AppserverID, cr.AppserverID)
	})
}

func TestQuerier_DeleteChannelRole(t *testing.T) {
	t.Run("Success:delete_channel_role", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		ch := testutil.TestChannel(t, nil, false)
		role := testutil.TestAppserverRole(t, &qx.AppserverRole{AppserverID: ch.AppserverID}, false)
		cr := testutil.TestChannelRole(t, &qx.ChannelRole{
			ChannelID: ch.ID, AppserverRoleID: role.ID, AppserverID: ch.AppserverID,
		}, false)

		// ACT
		count, err := qx.New(testutil.TestDbConn).DeleteChannelRole(ctx, cr.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Error:delete_nonexistent_channel_role", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		id := uuid.New()

		// ACT
		count, err := qx.New(testutil.TestDbConn).DeleteChannelRole(ctx, id)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestQuerier_GetChannelRoleById(t *testing.T) {
	t.Run("Success:get_channel_role_by_id", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		chRole := testutil.TestChannelRole(t, nil, false)

		// ACT
		found, err := qx.New(testutil.TestDbConn).GetChannelRoleById(ctx, chRole.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, chRole.ID, found.ID)
	})

	t.Run("Error:get_channel_role_by_id_not_found", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		id := uuid.New()

		// ACT
		_, err := qx.New(testutil.TestDbConn).GetChannelRoleById(ctx, id)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})
}

func TestQuerier_FilterChannelRole(t *testing.T) {
	t.Run("Success:filter_by_all_fields", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		cr := testutil.TestChannelRole(t, nil, false)

		params := qx.FilterChannelRoleParams{
			ChannelID:       pgtype.UUID{Bytes: cr.ChannelID, Valid: true},
			AppserverRoleID: pgtype.UUID{Bytes: cr.AppserverRoleID, Valid: true},
			AppserverID:     pgtype.UUID{Bytes: cr.AppserverID, Valid: true},
		}

		// ACT
		results, err := qx.New(testutil.TestDbConn).FilterChannelRole(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.Equal(t, cr.ID, results[0].ID)
	})

	t.Run("Success:filter_by_partial_fields", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		cr := testutil.TestChannelRole(t, nil, false)

		params := qx.FilterChannelRoleParams{
			ChannelID:       pgtype.UUID{Bytes: cr.ChannelID, Valid: true},
			AppserverRoleID: pgtype.UUID{Valid: false},
			AppserverID:     pgtype.UUID{Valid: false},
		}

		// ACT
		results, err := qx.New(testutil.TestDbConn).FilterChannelRole(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("EmptyResult:no_match_found", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})

		params := qx.FilterChannelRoleParams{
			ChannelID:       pgtype.UUID{Bytes: uuid.New(), Valid: true},
			AppserverRoleID: pgtype.UUID{Valid: false},
			AppserverID:     pgtype.UUID{Valid: false},
		}

		// ACT
		results, err := qx.New(testutil.TestDbConn).FilterChannelRole(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestQuerier_ListChannelRoles(t *testing.T) {
	t.Run("Success:list_channel_roles", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		cr := testutil.TestChannelRole(t, nil, false)

		// ACT
		results, err := qx.New(testutil.TestDbConn).ListChannelRoles(ctx, cr.ChannelID)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.Equal(t, cr.ChannelID, results[0].ChannelID)
	})

	t.Run("EmptyResult:list_channel_roles_empty", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		randomChannelID := uuid.New()

		// ACT
		results, err := qx.New(testutil.TestDbConn).ListChannelRoles(ctx, randomChannelID)

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}
