package permission

const (
	// Appserver Permission
	ManageChannels = 1 << 1
	ManageRoles     = 1 << 2
	ManageAppserver = 1 << 3

	// Channel Permissions

	// Sub Permissions
	ManageSubs = 1 << 1
)
