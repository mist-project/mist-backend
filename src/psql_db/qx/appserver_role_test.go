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
		ctx, _ := testutil.Setup(t, func() {})
		server := testutil.TestAppserver(t, nil, false)
		params := qx.CreateAppserverRoleParams{
			AppserverID:             server.ID,
			Name:                    "Admin",
			AppserverPermissionMask: 1,
			ChannelPermissionMask:   2,
			SubPermissionMask:       4,
		}

		// ACT
		role, err := qx.New(testutil.TestDbConn).CreateAppserverRole(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, params.Name, role.Name)
		assert.Equal(t, params.AppserverID, role.AppserverID)
	})
}

func TestQuerier_GetAppserverRoleById(t *testing.T) {
	t.Run("Success:get_appserver_role_by_id", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		role := testutil.TestAppserverRole(t, nil, false)

		// ACT
		result, err := qx.New(testutil.TestDbConn).GetAppserverRoleById(ctx, role.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, role.ID, result.ID)
	})

	t.Run("Error:role_does_not_exist", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		// ACT
		_, err := qx.New(testutil.TestDbConn).GetAppserverRoleById(ctx, uuid.New())

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})
}

func TestQuerier_DeleteAppserverRole(t *testing.T) {
	t.Run("Success:delete_appserver_role", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		role := testutil.TestAppserverRole(t, nil, false)

		// ACT
		count, err := qx.New(testutil.TestDbConn).DeleteAppserverRole(ctx, role.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Error:role_does_not_exist", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		// ACT
		count, err := qx.New(testutil.TestDbConn).DeleteAppserverRole(ctx, uuid.New())

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("Success:deleting_appserver_role_removes_associated_appserver_role_subs", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverSub(t, ctx, db)
		role := testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "test_role", AppserverID: su.Server.ID}, false)
		rolesub := testutil.TestAppserverRoleSub(t, &qx.AppserverRoleSub{
			AppuserID:       su.User.ID,
			AppserverID:     role.AppserverID,
			AppserverRoleID: role.ID,
			AppserverSubID:  su.Sub.ID,
		}, false)

		// ACT
		count, err := qx.New(testutil.TestDbConn).DeleteAppserverRole(ctx, role.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
		// Verify that the associated AppserverRoleSub is also deleted
		_, err = qx.New(testutil.TestDbConn).GetAppserverRoleSubById(ctx, rolesub.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})

	t.Run("Success:deleting_appserver_role_removes_associated_channel_roles", func(t *testing.T) {
		// ARRANGE
		ctx, db := testutil.Setup(t, func() {})
		su := factory.UserAppserverSub(t, ctx, db)
		role := testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "test_role", AppserverID: su.Server.ID}, false)
		channel := testutil.TestChannel(t, &qx.Channel{AppserverID: role.AppserverID, Name: "c1", IsPrivate: true}, false)
		channelRole := testutil.TestChannelRole(t, &qx.ChannelRole{
			AppserverID:     role.AppserverID,
			AppserverRoleID: role.ID,
			ChannelID:       channel.ID,
		}, false)

		// ACT
		count, err := qx.New(testutil.TestDbConn).DeleteAppserverRole(ctx, role.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// Verify that the associated AppserverRoleSub is also deleted
		_, err = qx.New(testutil.TestDbConn).GetAppserverRoleSubById(ctx, channelRole.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})

}

func TestQuerier_ListAppserverRoles(t *testing.T) {
	t.Run("Success:list_appserver_roles", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		role := testutil.TestAppserverRole(t, nil, false)

		// ACT
		roles, err := qx.New(testutil.TestDbConn).ListAppserverRoles(ctx, role.AppserverID)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, roles)
	})

	t.Run("Error:when_no_roles_exist", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		appserverID := uuid.New()

		// ACT
		roles, err := qx.New(testutil.TestDbConn).ListAppserverRoles(ctx, appserverID)

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
		role := testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "mod", AppserverID: su.Server.ID}, false)
		testutil.TestAppserverRoleSub(t, &qx.AppserverRoleSub{
			AppuserID: su.User.ID, AppserverID: su.Server.ID,
			AppserverSubID: su.Sub.ID, AppserverRoleID: role.ID,
		}, false)

		// ACT
		roles, err := qx.New(testutil.TestDbConn).GetAppuserRoles(ctx, qx.GetAppuserRolesParams{
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
		ctx, _ := testutil.Setup(t, func() {})
		nonexistentUserID := uuid.New()
		appserverID := uuid.New()

		// ACT
		roles, err := qx.New(testutil.TestDbConn).GetAppuserRoles(ctx, qx.GetAppuserRolesParams{
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
		role := testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "solo", AppserverID: su.Server.ID}, false)

		testutil.TestAppserverRoleSub(t, &qx.AppserverRoleSub{
			AppuserID: su.User.ID, AppserverRoleID: role.ID,
			AppserverID: su.Server.ID, AppserverSubID: su.Sub.ID,
		}, false)

		// ACT
		users, err := qx.New(testutil.TestDbConn).GetAppusersWithOnlySpecifiedRole(ctx, role.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, users)
		assert.Equal(t, su.User.ID, users[0].ID)
	})

	t.Run("Error:get_users_with_only_specified_role_no_users", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		role := testutil.TestAppserverRole(t, nil, false)

		// ACT
		users, err := qx.New(testutil.TestDbConn).GetAppusersWithOnlySpecifiedRole(ctx, role.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, users)
	})
}
