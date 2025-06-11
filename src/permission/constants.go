package permission

const (
	// Appserver Permission
	ManageChannels  = 1
	ManageRoles     = 1 << 1
	ManageAppserver = 1 << 2

	// Channel Permissions

	// Sub Permissions
	ManageSubs = 1
)
