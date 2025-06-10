package qx_test

import (
	"testing"

	"mist/src/psql_db/qx"
	"mist/src/testutil"
	"mist/src/testutil/factory"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestQuerier_CreateAppserverRole(t *testing.T) {
	t.Run("Success:create_appserver_role", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		server := factory.NewFactory(ctx, db).Appserver(t, 0, nil)
		params := qx.CreateAppserverRoleParams{
			AppserverID:             server.ID,
			Name:                    "Admin",
			AppserverPermissionMask: 1,
			ChannelPermissionMask:   2,
			SubPermissionMask:       4,
		}

		// ACT
		role, err := db.CreateAppserverRole(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, params.Name, role.Name)
		assert.Equal(t, params.AppserverID, role.AppserverID)
	})
}

func TestQuerier_GetAppserverRoleById(t *testing.T) {
	t.Run("Success:get_appserver_role_by_id", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		role := factory.NewFactory(ctx, db).AppserverRole(t, 0, nil)

		// ACT
		result, err := db.GetAppserverRoleById(ctx, role.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, role.ID, result.ID)
	})

	t.Run("Error:role_does_not_exist", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		// ACT
		_, err := db.GetAppserverRoleById(ctx, uuid.New())

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})
}

func TestQuerier_DeleteAppserverRole(t *testing.T) {
	t.Run("Success:delete_appserver_role", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		role := factory.NewFactory(ctx, db).AppserverRole(t, 0, nil)

		// ACT
		count, err := db.DeleteAppserverRole(ctx, role.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Error:role_does_not_exist", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})

		// ACT
		count, err := db.DeleteAppserverRole(ctx, uuid.New())

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("Success:deleting_appserver_role_removes_associated_appserver_role_subs", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		f := factory.NewFactory(ctx, db)
		f.AppserverSub(t, 0, nil)
		role := f.AppserverRole(t, 0, nil)
		roleSub := f.AppserverRoleSub(t, 0, nil)

		// ACT
		count, err := db.DeleteAppserverRole(ctx, role.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
		// Verify that the associated AppserverRoleSub is also deleted
		_, err = db.GetAppserverRoleSubById(ctx, roleSub.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})

	t.Run("Success:deleting_appserver_role_removes_associated_channel_roles", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		f := factory.NewFactory(ctx, db)
		sub := f.AppserverSub(t, 0, nil)
		role := f.AppserverRole(t, 0, nil)
		f.Channel(t, 0, &qx.Channel{Name: "foo", AppserverID: sub.AppserverID, IsPrivate: true})
		channelRole := f.ChannelRole(t, 0, nil)

		// ACT
		count, err := db.DeleteAppserverRole(ctx, role.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// Verify that the associated AppserverRoleSub is also deleted
		_, err = db.GetAppserverRoleSubById(ctx, channelRole.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})

}

func TestQuerier_ListAppserverRoles(t *testing.T) {
	t.Run("Success:list_appserver_roles", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		role := factory.NewFactory(ctx, db).AppserverRole(t, 0, nil)

		// ACT
		roles, err := db.ListAppserverRoles(ctx, role.AppserverID)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, roles)
	})

	t.Run("Error:when_no_roles_exist", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		appserverID := uuid.New()

		// ACT
		roles, err := db.ListAppserverRoles(ctx, appserverID)

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, roles)
	})
}

func TestQuerier_GetAppuserRoles(t *testing.T) {
	t.Run("Success:get_appuser_roles", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverSub(t, ctx, db)
		f := factory.NewFactory(ctx, db)
		role := f.AppserverRole(t, 0, nil)
		f.AppserverRoleSub(t, 0, nil)

		// ACT
		roles, err := db.GetAppuserRoles(ctx, qx.GetAppuserRolesParams{
			AppuserID:   su.User.ID,
			AppserverID: su.Server.ID,
		})

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, roles)
		assert.Equal(t, role.ID, roles[0].ID)
	})

	t.Run("Error:get_appuser_roles_for_nonexistent_user", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		nonexistentUserID := uuid.New()
		appserverID := uuid.New()

		// ACT
		roles, err := db.GetAppuserRoles(ctx, qx.GetAppuserRolesParams{
			AppuserID:   nonexistentUserID,
			AppserverID: appserverID,
		})

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, roles)
	})

}

func TestQuerier_GetAppusersWithOnlySpecifiedRole(t *testing.T) {
	t.Run("Success:get_users_with_only_specified_role", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverSub(t, ctx, db)
		f := factory.NewFactory(ctx, db)
		role := f.AppserverRole(t, 0, nil)
		factory.NewFactory(ctx, db).AppserverRoleSub(t, 0, nil)

		// ACT
		users, err := db.GetAppusersWithOnlySpecifiedRole(ctx, role.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, users)
		assert.Equal(t, su.User.ID, users[0].ID)
	})

	t.Run("Error:get_users_with_only_specified_role_no_users", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		role := factory.NewFactory(ctx, db).AppserverRole(t, 0, nil)

		// ACT
		users, err := db.GetAppusersWithOnlySpecifiedRole(ctx, role.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, users)
	})
}
