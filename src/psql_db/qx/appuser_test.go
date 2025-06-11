package qx_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"mist/src/psql_db/qx"
	"mist/src/testutil"
	"mist/src/testutil/factory"
)

func TestQuerier_CreateAppuser(t *testing.T) {
	t.Run("Success:create_appuser", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		id := uuid.New()
		username := "testuser"

		params := qx.CreateAppuserParams{
			ID:       id,
			Username: username,
		}

		// ACT
		user, err := db.CreateAppuser(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, id, user.ID)
		assert.Equal(t, username, user.Username)
	})
}

func TestQuerier_GetAppuserById(t *testing.T) {
	t.Run("Success:get_appuser_by_id", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		u := factory.NewFactory(ctx, db).Appuser(t, 0, nil)

		// ACT
		user, err := db.GetAppuserById(ctx, u.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, user.ID, user.ID)
		assert.Equal(t, user.Username, user.Username)
	})

	t.Run("Error:get_nonexistent_appuser", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		nonexistentID := uuid.New()

		// ACT
		_, err := db.GetAppuserById(ctx, nonexistentID)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})
}
