package qx_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"mist/src/psql_db/qx"
	"mist/src/testutil"
)

func TestQuerier_CreateAppuser(t *testing.T) {
	t.Run("Success:create_appuser", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		id := uuid.New()
		username := "testuser"

		params := qx.CreateAppuserParams{
			ID:       id,
			Username: username,
		}

		// ACT
		user, err := qx.New(testutil.TestDbConn).CreateAppuser(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, id, user.ID)
		assert.Equal(t, username, user.Username)
	})
}

func TestQuerier_GetAppuserById(t *testing.T) {
	t.Run("Success:get_appuser_by_id", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		id := uuid.New()
		username := "retrievable_user"
		_, err := qx.New(testutil.TestDbConn).CreateAppuser(ctx, qx.CreateAppuserParams{
			ID:       id,
			Username: username,
		})
		assert.NoError(t, err)

		// ACT
		user, err := qx.New(testutil.TestDbConn).GetAppuserById(ctx, id)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, id, user.ID)
		assert.Equal(t, username, user.Username)
	})

	t.Run("Error:get_nonexistent_appuser", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		nonexistentID := uuid.New()

		// ACT
		_, err := qx.New(testutil.TestDbConn).GetAppuserById(ctx, nonexistentID)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})
}
