package repository

import (
	"context"

	"github.com/ichzzy/go_o11y_base/internal/domain"
	"github.com/ichzzy/go_o11y_base/internal/domain/entity"
)

type RbacRepository interface {
	GetPolicyRules(ctx context.Context) ([]entity.PolicyRule, error)
	ListRoles(ctx context.Context, page domain.PageReq) ([]entity.Role, int64, error)
	GetRoleByID(ctx context.Context, roleID uint64) (*entity.Role, error)
	CreateRoleWithPermissions(ctx context.Context, role *entity.Role, permissionIDs []uint64) error
	UpdateRoleWithPermissions(ctx context.Context, param *UpdateRoleParam) error
	GetRoleWithPermissions(ctx context.Context, roleID uint64) (*entity.RoleWithPermissions, error)
	ListPermissions(ctx context.Context, page domain.PageReq) ([]entity.Permission, int64, error)
	GetRolePermissionIDs(ctx context.Context, roleID uint64) ([]uint64, error)
	GetUserVisibleMenus(ctx context.Context, userID uint64) ([]entity.Permission, error)
	GetRoleIDByUserID(ctx context.Context, userID uint64) (uint64, error)
}

type UpdateRoleParam struct {
	ID            uint64
	Name          *string
	PermissionIDs *[]uint64
}
