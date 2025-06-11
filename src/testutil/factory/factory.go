package factory

import (
	"context"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
	"testing"
)

type Factory struct {
	ctx context.Context
	db  db.Querier
}

func NewFactory(ctx context.Context, db db.Querier) *Factory {
	return &Factory{
		ctx: ctx,
		db:  db,
	}
}

func (f *Factory) Appuser(t *testing.T, index int, appuser *qx.Appuser) *qx.Appuser {
	if index < 0 || index >= len(fakeAppusers) {
		t.Fatalf("Invalid factory index: %d", index)
	}

	var (
		u   qx.Appuser
		err error
	)

	if appuser != nil {
		// with provided appuser, create using that one
		u, err = f.db.GetAppuserById(f.ctx, appuser.ID)

		if err == nil {
			// if appuser exists, return it else create a new one
			return &u
		}

		u, err = f.db.CreateAppuser(
			f.ctx, qx.CreateAppuserParams{
				ID:       appuser.ID,
				Username: appuser.Username,
			},
		)

	} else {
		// create a new appuser using the fake data

		u, err = f.db.GetAppuserById(f.ctx, fakeAppusers[index].ID)
		if err == nil {
			// if appuser exists, return it else create a new one
			return &u
		}

		u, err = f.db.CreateAppuser(
			f.ctx, qx.CreateAppuserParams{
				ID:       fakeAppusers[index].ID,
				Username: fakeAppusers[index].Username,
			},
		)
	}

	if err != nil {
		t.Fatalf("Unable to create appuser. Error: %v", err)
	}

	// cache the created appuser for future use
	fakeAppusers[index].ID = u.ID

	return &u
}

func (f *Factory) Appserver(t *testing.T, index int, appserver *qx.Appserver) *qx.Appserver {

	if index < 0 || index >= len(fakeAppservers) {
		t.Fatalf("Invalid factory index: %d", index)
	}

	var (
		s   qx.Appserver
		err error
	)

	if appserver != nil {
		// with provided appserver, create using that one
		s, err = f.db.GetAppserverById(f.ctx, appserver.ID)

		if err == nil {
			// if appserver exists, return it else create a new one
			return &s
		}

		s, err = f.db.CreateAppserver(
			f.ctx, qx.CreateAppserverParams{
				Name:      appserver.Name,
				AppuserID: appserver.AppuserID,
			},
		)

	} else {
		u := f.Appuser(t, index, nil)

		s, err = f.db.GetAppserverById(f.ctx, fakeAppservers[index].ID)
		if err == nil {
			// if appserver exists, return it else create a new one
			return &s
		}

		s, err = f.db.CreateAppserver(
			f.ctx, qx.CreateAppserverParams{
				Name:      fakeAppservers[index].Name,
				AppuserID: u.ID,
			},
		)
	}

	if err != nil {
		t.Fatalf("Unable to create appserver. Error: %v", err)
	}

	// cache the created appserver for future use
	fakeAppservers[index].ID = s.ID

	return &s
}

// AppserverSub creates or retrieves an appserver subscription by index.
func (f *Factory) AppserverSub(t *testing.T, index int, appserverSub *qx.AppserverSub) *qx.AppserverSub {
	if index < 0 || index >= len(fakeAppserverSubs) {
		t.Fatalf("Invalid factory index: %d", index)
	}

	var (
		sub qx.AppserverSub
		err error
	)

	if appserverSub != nil {
		// with provided appserverSub, create using that one
		sub, err = f.db.GetAppserverSubById(f.ctx, appserverSub.ID)

		if err == nil {
			// if appserverSub exists, return it else create a new one
			return &sub
		}

		sub, err = f.db.CreateAppserverSub(
			f.ctx, qx.CreateAppserverSubParams{
				AppserverID: appserverSub.AppserverID,
				AppuserID:   appserverSub.AppuserID,
			},
		)

	} else {
		sub, err = f.db.GetAppserverSubById(f.ctx, fakeAppserverSubs[index].ID)
		if err == nil {
			// if appserverSub exists, return it else create a new one
			return &sub
		}

		s := f.Appserver(t, index, nil)
		user := f.Appuser(t, index, nil)
		sub, err = f.db.CreateAppserverSub(
			f.ctx, qx.CreateAppserverSubParams{
				AppserverID: s.ID,
				AppuserID:   user.ID,
			},
		)
	}

	if err != nil {
		t.Fatalf("Unable to create appserver sub. Error: %v", err)
	}

	// cache the created appserver sub for future use
	fakeAppserverSubs[index].ID = sub.ID

	return &sub
}

// AppserverRole creates or retrieves an appserver role by index.
func (f *Factory) AppserverRole(t *testing.T, index int, appserverRole *qx.AppserverRole) *qx.AppserverRole {
	if index < 0 || index >= len(fakeAppserverRoles) {
		t.Fatalf("Invalid factory index: %d", index)
	}

	var (
		role qx.AppserverRole
		err  error
	)

	if appserverRole != nil {

		// with provided appserverRole, create using that one
		role, err = f.db.GetAppserverRoleById(f.ctx, appserverRole.ID)

		if err == nil {
			// if appserverRole exists, return it else create a new one
			return &role
		}

		role, err = f.db.CreateAppserverRole(
			f.ctx, qx.CreateAppserverRoleParams{
				Name:                    appserverRole.Name,
				AppserverID:             appserverRole.AppserverID,
				AppserverPermissionMask: appserverRole.AppserverPermissionMask,
				ChannelPermissionMask:   appserverRole.ChannelPermissionMask,
				SubPermissionMask:       appserverRole.SubPermissionMask,
			},
		)
	} else {

		role, err = f.db.GetAppserverRoleById(f.ctx, fakeAppserverRoles[index].ID)
		if err == nil {
			// if appserverRole exists, return it else create a new one
			return &role
		}

		s := f.Appserver(t, index, nil)
		role, err = f.db.CreateAppserverRole(
			f.ctx, qx.CreateAppserverRoleParams{
				Name:                    fakeAppserverRoles[index].Name,
				AppserverID:             s.ID,
				AppserverPermissionMask: fakeAppserverRoles[index].AppserverPermissionMask,
				ChannelPermissionMask:   fakeAppserverRoles[index].ChannelPermissionMask,
				SubPermissionMask:       fakeAppserverRoles[index].SubPermissionMask,
			},
		)
	}

	if err != nil {
		t.Fatalf("Unable to create appserver role. Error: %v", err)
	}

	// cache the created appserver role for future use
	fakeAppserverRoles[index].ID = role.ID

	return &role
}

// reate AppserverroleSub creates or retrieves an appserver role subscription by index.
func (f *Factory) AppserverRoleSub(t *testing.T, index int, appserverRoleSub *qx.AppserverRoleSub) *qx.AppserverRoleSub {
	if index < 0 || index >= len(fakeAppserverRoleSubs) {
		t.Fatalf("Invalid factory index: %d", index)
	}

	var (
		roleSub qx.AppserverRoleSub
		err     error
	)

	if appserverRoleSub != nil {
		// with provided appserverRoleSub, create using that one
		roleSub, err = f.db.GetAppserverRoleSubById(f.ctx, appserverRoleSub.ID)
		if err == nil {
			// if appserverRoleSub exists, return it else create a new one
			return &roleSub
		}

		roleSub, err = f.db.CreateAppserverRoleSub(
			f.ctx, qx.CreateAppserverRoleSubParams{
				AppserverRoleID: appserverRoleSub.AppserverRoleID,
				AppuserID:       appserverRoleSub.AppuserID,
				AppserverSubID:  appserverRoleSub.AppserverSubID,
				AppserverID:     appserverRoleSub.AppserverID,
			},
		)

	} else {
		roleSub, err = f.db.GetAppserverRoleSubById(f.ctx, fakeAppserverRoleSubs[index].ID)
		if err == nil {
			// if appserverRoleSub exists, return it else create a new one
			return &roleSub
		}
		s := f.Appserver(t, index, nil)
		user := f.Appuser(t, index, nil)
		sub := f.AppserverSub(t, index, nil)
		role := f.AppserverRole(t, index, nil)
		roleSub, err = f.db.CreateAppserverRoleSub(
			f.ctx, qx.CreateAppserverRoleSubParams{
				AppserverRoleID: role.ID,
				AppuserID:       user.ID,
				AppserverSubID:  sub.ID,
				AppserverID:     s.ID,
			},
		)
	}

	if err != nil {
		t.Fatalf("Unable to create appserver role sub. Error: %v", err)
	}

	// cache the created appserver role sub for future use
	fakeAppserverRoleSubs[index].ID = roleSub.ID
	return &roleSub
}

// Channel creates or retrieves a channel by index.
func (f *Factory) Channel(t *testing.T, index int, channel *qx.Channel) *qx.Channel {
	var (
		c   qx.Channel
		ch  qx.Channel
		err error
	)

	if index < 0 || index >= len(fakeChannels) {
		t.Fatalf("Invalid factory index: %d", index)
	}

	if channel != nil {
		// with provided channel, create using that one
		c, err := f.db.GetChannelById(f.ctx, fakeChannels[index].ID)
		if err == nil {
			// if channel exists, return it else create a new one
			return &c
		}

		ch, err = f.db.CreateChannel(
			f.ctx, qx.CreateChannelParams{
				Name:        channel.Name,
				AppserverID: channel.AppserverID,
				IsPrivate:   channel.IsPrivate,
			},
		)
	} else {

		c, err = f.db.GetChannelById(f.ctx, fakeChannels[index].ID)

		if err == nil {
			// if channel exists, return it else create a new one
			return &c
		}

		s := f.Appserver(t, index, nil)

		ch, err = f.db.CreateChannel(
			f.ctx, qx.CreateChannelParams{
				Name:        c.Name,
				AppserverID: s.ID,
				IsPrivate:   c.IsPrivate,
			},
		)
		fakeChannels[index].ID = ch.ID
	}

	if err != nil {
		t.Fatalf("Unable to create channel. Error: %v", err)
	}

	// cache the created channel for future use
	fakeChannels[index].ID = ch.ID

	return &ch
}

func (f *Factory) ChannelRole(t *testing.T, index int, channelRole *qx.ChannelRole) *qx.ChannelRole {
	var (
		cr  qx.ChannelRole
		err error
	)

	if index < 0 || index >= len(fakeChannelRoles) {
		t.Fatalf("Invalid factory index: %d", index)
	}

	if channelRole != nil {
		// with provided channelRole, create using that one
		cr, err = f.db.GetChannelRoleById(f.ctx, channelRole.ID)
		if err == nil {
			// if channelRole exists, return it else create a new one
			return &cr
		}

		cr, err = f.db.CreateChannelRole(
			f.ctx, qx.CreateChannelRoleParams{
				ChannelID:       channelRole.ChannelID,
				AppserverID:     channelRole.AppserverID,
				AppserverRoleID: channelRole.AppserverRoleID,
			},
		)
	} else {

		cr, err = f.db.GetChannelRoleById(f.ctx, fakeChannelRoles[index].ID)

		if err == nil {
			// if channel exists, return it else create a new one
			return &cr
		}

		s := f.Appserver(t, index, nil)
		ch := f.Channel(t, index, nil)
		role := f.AppserverRole(t, index, nil)

		cr, err = f.db.CreateChannelRole(
			f.ctx, qx.CreateChannelRoleParams{
				ChannelID:       ch.ID,
				AppserverID:     s.ID,
				AppserverRoleID: role.ID,
			},
		)
	}

	if err != nil {
		t.Fatalf("Unable to create channel role. Error: %v", err)
	}

	fakeChannelRoles[index].ID = cr.ID

	return &cr
}
