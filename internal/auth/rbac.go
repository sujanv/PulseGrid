package auth

type Role string

const (
	RoleViewer   Role = "viewer"
	RoleOperator Role = "operator"
	RoleAdmin    Role = "admin"
	RoleOwner    Role = "owner"
)

type Permission string

const (
	PermTargetsRead       Permission = "targets:read"
	PermTargetsWrite      Permission = "targets:write"
	PermAlertsManage      Permission = "alerts:manage"
	PermRemediationExec   Permission = "remediation:execute"
	PermSLOAdmin          Permission = "slo:admin"
	PermAdminManage       Permission = "admin:manage"
)

var rolePermissions = map[Role][]Permission{
	RoleViewer: {
		PermTargetsRead,
	},
	RoleOperator: {
		PermTargetsRead,
		PermTargetsWrite,
		PermRemediationExec,
	},
	RoleAdmin: {
		PermTargetsRead,
		PermTargetsWrite,
		PermAlertsManage,
		PermRemediationExec,
		PermSLOAdmin,
	},
	RoleOwner: {
		PermTargetsRead,
		PermTargetsWrite,
		PermAlertsManage,
		PermRemediationExec,
		PermSLOAdmin,
		PermAdminManage,
	},
}

type RBACChecker struct{}

func NewRBACChecker() *RBACChecker {
	return &RBACChecker{}
}

func (r *RBACChecker) HasPermission(userRole Role, perm Permission) bool {
	perms, ok := rolePermissions[userRole]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == perm || p == PermAdminManage {
			return true
		}
	}
	return false
}

func (r *RBACChecker) IsAtLeastRole(userRole, requiredRole Role) bool {
	hierarchy := map[Role]int{
		RoleViewer:   1,
		RoleOperator: 2,
		RoleAdmin:    3,
		RoleOwner:    4,
	}
	return hierarchy[userRole] >= hierarchy[requiredRole]
}
