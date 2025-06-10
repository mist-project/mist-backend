package factory

import (
	"mist/src/psql_db/qx"
	"mist/src/testutil"

	"github.com/google/uuid"
)

var (

	// fake appusers
	fakeAppusers = []*qx.Appuser{
		{ID: uuid.MustParse("874fb22c-8e3a-4504-9a4b-437c22655865"), Username: "testuser0"},
		{ID: uuid.MustParse("874fb22c-8e3a-4504-9a4b-437c22655866"), Username: "testuser1"},
		{ID: uuid.MustParse("874fb22c-8e3a-4504-9a4b-437c22655867"), Username: "testuser2"},
		{ID: uuid.MustParse("874fb22c-8e3a-4504-9a4b-437c22655868"), Username: "testuser3"},
		{ID: uuid.MustParse("874fb22c-8e3a-4504-9a4b-437c22655869"), Username: "testuser4"},
	}

	// fake appservers
	fakeAppservers = []*qx.Appserver{
		{Name: "testserver0", AppuserID: uuid.MustParse(testutil.DefaultUserId)},
		{Name: "testserver1", AppuserID: uuid.MustParse(testutil.DefaultUserId)},
		{Name: "testserver2", AppuserID: uuid.MustParse(testutil.DefaultUserId)},
		{Name: "testserver3", AppuserID: uuid.MustParse(testutil.DefaultUserId)},
		{Name: "testserver4", AppuserID: uuid.MustParse(testutil.DefaultUserId)},
	}

	// fake appserver subs
	fakeAppserverSubs = []*qx.AppserverSub{
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
	}

	// fake appserver roles
	fakeAppserverRoles = []*qx.AppserverRole{
		{ID: uuid.New(), Name: "role0", AppserverPermissionMask: int64(0), ChannelPermissionMask: int64(0), SubPermissionMask: int64(0)},
		{ID: uuid.New(), Name: "role1", AppserverPermissionMask: int64(0), ChannelPermissionMask: int64(0), SubPermissionMask: int64(0)},
		{ID: uuid.New(), Name: "role2", AppserverPermissionMask: int64(0), ChannelPermissionMask: int64(0), SubPermissionMask: int64(0)},
		{ID: uuid.New(), Name: "role3", AppserverPermissionMask: int64(0), ChannelPermissionMask: int64(0), SubPermissionMask: int64(0)},
		{ID: uuid.New(), Name: "role4", AppserverPermissionMask: int64(0), ChannelPermissionMask: int64(0), SubPermissionMask: int64(0)},
	}

	// fake appserver role subs
	fakeAppserverRoleSubs = []*qx.AppserverRoleSub{
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
	}

	// fake channels
	fakeChannels = []*qx.Channel{
		{ID: uuid.New(), Name: "testchannel0", IsPrivate: false},
		{ID: uuid.New(), Name: "testchannel1", IsPrivate: true},
		{ID: uuid.New(), Name: "testchannel2", IsPrivate: false},
		{ID: uuid.New(), Name: "testchannel3", IsPrivate: true},
		{ID: uuid.New(), Name: "testchannel4", IsPrivate: false},
	}

	// fake channel roles
	fakeChannelRoles = []*qx.ChannelRole{
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
	}
)
