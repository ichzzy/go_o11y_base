package usecase

import (
	"context"

	"github.com/ichzzy/go_o11y_base/internal/domain"
	"github.com/ichzzy/go_o11y_base/internal/domain/entity"
)

type RbacUsecase interface {
	LoadPolicy(ctx context.Context) error
	Enforce(sub, obj, act string) (bool, error)

	ListRoles(ctx context.Context, page domain.PageReq) ([]entity.Role, int64, error)
	CreateRoleWithPermissions(ctx context.Context, name string, permissionIDs []uint64) error
	UpdateRoleWithPermissions(ctx context.Context, id uint64, name string, permissionIDs []uint64) error
	GetRoleWithPermissions(ctx context.Context, roleID uint64) (*entity.RoleWithPermissions, error)
	ListPermissions(ctx context.Context, page domain.PageReq) ([]entity.Permission, int64, error)
	GetUserVisibleMenus(ctx context.Context, userID uint64) ([]entity.Permission, error)
}
