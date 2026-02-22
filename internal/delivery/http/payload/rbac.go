package payload

import "github.com/ichzzy/go_o11y_base/internal/domain"

type ListRolesReq struct {
	domain.PageReq
}

type CreateRoleReq struct {
	Name          string   `json:"name"`
	PermissionIDs []uint64 `json:"permission_ids"`
}

type UpdateRoleReq struct {
	Name          string   `json:"name"`
	PermissionIDs []uint64 `json:"permission_ids"`
}

type ListPermissionsReq struct {
	domain.PageReq
}
