package role

import "github.com/COMTOP1/AFC-GO/role"

// SufficientPermissionsFor takes a permission for a task and returns that permission and higher permissions that would be acceptable
func SufficientPermissionsFor(r1 role.Role) map[string]bool {
	_ = r1
	return nil
}
