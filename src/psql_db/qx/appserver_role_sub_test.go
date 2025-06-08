package qx_test

import (
	"mist/src/psql_db/qx"
	"mist/src/testutil"
	"mist/src/testutil/factory"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestQuerier_ListServerRoleSubs(t *testing.T) {
	t.Run("Successlist_role_subs_by_appserver", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		role1 := testutil.TestAppserverRoleSub(t, nil, false)
		testutil.TestAppserverRoleSub(t, nil, false)

		// ACT
		results, err := qx.New(testutil.TestDbConn).ListServerRoleSubs(ctx, role1.AppserverID)

		// ASSERT
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, role1.AppserverID, results[0].AppserverID)
	})
}

func TestQuerier_CreateAppserverRoleSub(t *testing.T) {
	t.Run("Successcreate_appserver_role_sub", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		su := factory.UserAppserverSub(t)
		role1 := testutil.TestAppserverRole(t, &qx.AppserverRole{Name: "foo", AppserverID: su.Server.ID}, false)

		// ACT
		r, err := qx.New(testutil.TestDbConn).CreateAppserverRoleSub(ctx, qx.CreateAppserverRoleSubParams{
			AppserverSubID:  su.Sub.ID,
			AppserverRoleID: role1.ID,
			AppuserID:       su.User.ID,
			AppserverID:     su.Server.ID,
		})

		// ASSERT
		assert.NoError(t, err)
		assert.NotNil(t, r)
	})

	t.Run("Error:when_appserver_sub_and_appserver_role_dont_belong_to_same_server_it_errors", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		user := testutil.TestAppuser(t, nil, false)
		su := factory.UserAppserverSub(t)
		role1 := testutil.TestAppserverRole(t, nil, false)

		// ACT
		_, err := qx.New(testutil.TestDbConn).CreateAppserverRoleSub(ctx, qx.CreateAppserverRoleSubParams{
			AppserverSubID:  su.Sub.ID,
			AppserverRoleID: role1.ID,
			AppuserID:       user.ID,
			AppserverID:     role1.AppserverID,
		})

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `violates foreign key constraint "appserver_role_sub_uk_server_and_sub"`)
	})
}

func TestQuerier_GetById(t *testing.T) {
	t.Run("Successget_appserver_role_sub_by_id", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		role1 := testutil.TestAppserverRoleSub(t, nil, false)

		// ACT
		result, err := qx.New(testutil.TestDbConn).GetAppserverRoleSubById(ctx, role1.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, role1.ID, result.ID)
	})

	t.Run("Error:when_appserver_role_sub_does_not_exist_it_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		id := uuid.New()

		// ACT
		_, err := qx.New(testutil.TestDbConn).GetAppserverRoleSubById(ctx, id)

		// ASSERT
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows in result set")
	})

}

func TestQuerier_DeleteAppserverRoleSub(t *testing.T) {
	t.Run("Successdelete_appserver_role_sub", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		role1 := testutil.TestAppserverRoleSub(t, nil, false)

		// ACT
		count, err := qx.New(testutil.TestDbConn).DeleteAppserverRoleSub(ctx, role1.ID)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Error:when_appserver_role_sub_does_not_exist_it_returns_zero_count", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		id := uuid.New()

		// ACT
		count, _ := qx.New(testutil.TestDbConn).DeleteAppserverRoleSub(ctx, id)

		// ASSERT
		assert.Equal(t, int64(0), count)
	})
}

func TestQuerier_FilterAppserverRoleSub(t *testing.T) {
	t.Run("Successfilter_by_all_fields", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		roleSub := testutil.TestAppserverRoleSub(t, nil, false)

		params := qx.FilterAppserverRoleSubParams{
			AppuserID:       pgtype.UUID{Bytes: roleSub.AppuserID, Valid: true},
			AppserverID:     pgtype.UUID{Bytes: roleSub.AppserverID, Valid: true},
			AppserverRoleID: pgtype.UUID{Bytes: roleSub.AppserverRoleID, Valid: true},
			AppserverSubID:  pgtype.UUID{Bytes: roleSub.AppserverSubID, Valid: true},
		}

		// ACT
		results, err := qx.New(testutil.TestDbConn).FilterAppserverRoleSub(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.Equal(t, roleSub.ID, results[0].ID)
	})

	t.Run("Successfilter_by_partial_fields", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		roleSub := testutil.TestAppserverRoleSub(t, nil, false)

		params := qx.FilterAppserverRoleSubParams{
			AppuserID:       pgtype.UUID{Bytes: roleSub.AppuserID, Valid: true},
			AppserverID:     pgtype.UUID{Valid: false},
			AppserverRoleID: pgtype.UUID{Valid: false},
			AppserverSubID:  pgtype.UUID{Valid: false},
		}

		// ACT
		results, err := qx.New(testutil.TestDbConn).FilterAppserverRoleSub(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("EmptyResult:when_filter_does_not_match", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		_ = testutil.TestAppserverRoleSub(t, nil, false)

		params := qx.FilterAppserverRoleSubParams{
			AppuserID:       pgtype.UUID{Bytes: uuid.New(), Valid: true},
			AppserverID:     pgtype.UUID{Valid: false},
			AppserverRoleID: pgtype.UUID{Valid: false},
			AppserverSubID:  pgtype.UUID{Valid: false},
		}

		// ACT
		results, err := qx.New(testutil.TestDbConn).FilterAppserverRoleSub(ctx, params)

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}
