package factory

import (
	"mist/src/permission"
	"mist/src/psql_db/qx"
	"mist/src/testutil"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

type testUser struct {
	User   *qx.Appuser
	Server *qx.Appserver
	Sub    *qx.AppserverSub
}

// Returns a testUser that owns a server.
func UserAppserverOwner(t *testing.T) *testUser {
	var (
		u   *qx.Appuser
		s   *qx.Appserver
		sub *qx.AppserverSub
	)
	u = testutil.TestAppuser(t, nil, true)
	s = testutil.TestAppserver(t, nil, true)
	sub = testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: s.ID, AppuserID: u.ID}, false)
	return &testUser{User: u, Server: s, Sub: sub}
}

// Returns a testUser that is subscribed to a server.
func UserAppserverSub(t *testing.T) *testUser {
	var (
		u   *qx.Appuser
		s   *qx.Appserver
		sub *qx.AppserverSub
	)
	u = testutil.TestAppuser(t, nil, true)
	s = testutil.TestAppserver(t, nil, false)
	sub = testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: s.ID, AppuserID: u.ID}, false)
	return &testUser{User: u, Server: s, Sub: sub}
}

// Returns a testUser with random server sub (the server in testUser object does not match the sub's server).
func UserAppserverUnsub(t *testing.T) *testUser {
	var (
		// u   *qx.Appuser
		u   *qx.Appuser
		s   *qx.Appserver
		s2  *qx.Appserver
		sub *qx.AppserverSub
	)
	u = testutil.TestAppuser(t, nil, false)
	s = testutil.TestAppserver(t, nil, false)
	s2 = testutil.TestAppserver(t, nil, false)
	sub = testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: s2.ID, AppuserID: u.ID}, false)

	return &testUser{User: u, Server: s, Sub: sub}
}

// Returns a testUser with server permission.
func UserAppserverWithPermission(t *testing.T) *testUser {
	var (
		u   *qx.Appuser
		s   *qx.Appserver
		sub *qx.AppserverSub
	)
	u = testutil.TestAppuser(t, nil, true)
	s = testutil.TestAppserver(t, nil, false)
	sub = testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: s.ID, AppuserID: u.ID}, false)

	testutil.TestAppserverPermission(t, &qx.AppserverPermission{
		AppserverID: sub.AppserverID,
		AppuserID:   u.ID,
		ReadAll:     pgtype.Bool{Valid: true, Bool: true},
		WriteAll:    pgtype.Bool{Valid: true, Bool: true},
		DeleteAll:   pgtype.Bool{Valid: true, Bool: true},
	}, false)

	return &testUser{User: u, Server: s, Sub: sub}
}

// Returns a testUser with all permissions.
func UserAppserverWithAllPermissions(t *testing.T) *testUser {
	var (
		u   *qx.Appuser
		s   *qx.Appserver
		sub *qx.AppserverSub
	)
	u = testutil.TestAppuser(t, nil, true)
	s = testutil.TestAppserver(t, nil, false)
	sub = testutil.TestAppserverSub(t, &qx.AppserverSub{AppserverID: s.ID, AppuserID: u.ID}, false)

	role := testutil.TestAppserverRole(
		t,
		&qx.AppserverRole{
			AppserverID:             s.ID,
			Name:                    "admin",
			AppserverPermissionMask: permission.ManageAppserver | permission.ManageRoles | permission.ManagedChannels,
			ChannelPermissionMask:   0,
			SubPermissionMask:       permission.ManageSubs,
		},
		false,
	)

	testutil.TestAppserverRoleSub(t, &qx.AppserverRoleSub{
		AppserverID:     s.ID,
		AppuserID:       u.ID,
		AppserverSubID:  sub.ID,
		AppserverRoleID: role.ID,
	}, false)

	return &testUser{User: u, Server: s, Sub: sub}
}

// func TestAppserverRole(t *testing.T, aRole *qx.AppserverRole, base bool) *qx.AppserverRole {
