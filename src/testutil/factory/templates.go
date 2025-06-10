package factory

import (
	"context"
	"mist/src/permission"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
	"mist/src/testutil"
	"testing"

	"github.com/google/uuid"
)

type testUser struct {
	User   *qx.Appuser
	Server *qx.Appserver
	Sub    *qx.AppserverSub
}

func UserAppserverOwner(t *testing.T, ctx context.Context, db db.Querier) *testUser {
	var (
		u   *qx.Appuser
		s   *qx.Appserver
		sub *qx.AppserverSub
	)

	f := NewFactory(ctx, db)

	parsedUid, err := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))

	if err != nil {
		t.Fatalf("failed to parse user ID from context: %v", err)
	}

	u = f.Appuser(t, 0, &qx.Appuser{ID: parsedUid, Username: "testuser"})
	s = f.Appserver(t, 0, &qx.Appserver{Name: "testserver", AppuserID: u.ID})
	sub = f.AppserverSub(t, 0, &qx.AppserverSub{AppserverID: s.ID, AppuserID: u.ID})

	return &testUser{User: u, Server: s, Sub: sub}
}

// Returns a testUser that is subscribed to a server.
func UserAppserverSub(t *testing.T, ctx context.Context, db db.Querier) *testUser {
	var (
		u   *qx.Appuser
		s   *qx.Appserver
		sub *qx.AppserverSub
	)
	f := NewFactory(ctx, db)

	parsedUid, err := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))

	if err != nil {
		t.Fatalf("failed to parse user ID from context: %v", err)
	}

	u = f.Appuser(t, 0, &qx.Appuser{ID: parsedUid, Username: "testuser"})
	s = f.Appserver(t, 0, nil)
	sub = f.AppserverSub(t, 0, nil)

	return &testUser{User: u, Server: s, Sub: sub}
}

// Returns a testUser with random server sub (the server in testUser object does not match the sub's server).
func UserAppserverUnsub(t *testing.T, ctx context.Context, db db.Querier) *testUser {
	var (
		// u   *qx.Appuser
		u   *qx.Appuser
		s   *qx.Appserver
		s1  *qx.Appserver
		sub *qx.AppserverSub
	)

	f := NewFactory(ctx, db)
	parsedUid, err := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))

	if err != nil {
		t.Fatalf("failed to parse user ID from context: %v", err)
	}

	u = f.Appuser(t, 0, &qx.Appuser{ID: parsedUid, Username: "testuser"})
	owner := f.Appuser(t, 1, &qx.Appuser{ID: uuid.New(), Username: "owner"})
	s = f.Appserver(t, 0, nil)
	s1 = f.Appserver(t, 1, nil)
	sub = f.AppserverSub(t, 0, &qx.AppserverSub{AppserverID: s1.ID, AppuserID: owner.ID})

	return &testUser{User: u, Server: s, Sub: sub}
}

// Returns a testUser with all permissions.
func UserAppserverWithAllPermissions(t *testing.T, ctx context.Context, db db.Querier) *testUser {
	var (
		u   *qx.Appuser
		s   *qx.Appserver
		sub *qx.AppserverSub
	)

	f := NewFactory(ctx, db)

	// Create a user, server and subscription
	parsedUid, err := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))

	if err != nil {
		t.Fatalf("failed to parse user ID from context: %v", err)
	}

	u = f.Appuser(t, 0, &qx.Appuser{ID: parsedUid, Username: "testuser"})
	s = f.Appserver(t, 0, nil)
	sub = f.AppserverSub(t, 0, &qx.AppserverSub{AppserverID: s.ID, AppuserID: u.ID})

	role := f.AppserverRole(
		t,
		0,
		&qx.AppserverRole{
			AppserverID:             s.ID,
			Name:                    "admin",
			AppserverPermissionMask: permission.ManageAppserver | permission.ManageRoles | permission.ManageChannels,
			ChannelPermissionMask:   0,
			SubPermissionMask:       permission.ManageSubs,
		},
	)

	f.AppserverRoleSub(t, 0, &qx.AppserverRoleSub{
		AppserverID:     s.ID,
		AppuserID:       u.ID,
		AppserverSubID:  sub.ID,
		AppserverRoleID: role.ID,
	})

	return &testUser{User: u, Server: s, Sub: sub}
}
