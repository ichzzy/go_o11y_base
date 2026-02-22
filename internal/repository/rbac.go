package repository

import (
	"context"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/ichzzy/go_o11y_base/internal/domain"
	"github.com/ichzzy/go_o11y_base/internal/domain/entity"
	domain_repo "github.com/ichzzy/go_o11y_base/internal/domain/repository"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

type RbacRepoParam struct {
	dig.In

	DB *gorm.DB `name:"mysql"`
}

type rbacRepository struct {
	db *gorm.DB
}

func NewRbacRepository(param RbacRepoParam) domain_repo.RbacRepository {
	return &rbacRepository{db: param.DB}
}

func (r *rbacRepository) GetPolicyRules(
	ctx context.Context,
) ([]entity.PolicyRule, error) {
	var rules []entity.PolicyRule
	err := r.db.WithContext(ctx).Table("role_permission rp").
		Select("r.id as role_id, p.http_path as path, p.http_method as method").
		Joins("JOIN role r ON rp.role_id = r.id").
		Joins("JOIN permission p ON rp.permission_id = p.id").
		Where("p.type = ?", entity.PermissionTypeAPI).
		Scan(&rules).Error
	if err != nil {
		return nil, domain.ErrInternal.ReWrap(err)
	}
	return rules, nil
}

func (r *rbacRepository) ListRoles(
	ctx context.Context,
	page domain.PageReq,
) ([]entity.Role, int64, error) {
	var roles []entity.Role
	var total int64

	db := r.db.WithContext(ctx)
	if err := db.Model(&entity.Role{}).Count(&total).Error; err != nil {
		return nil, 0, domain.ErrInternal.ReWrap(err)
	}

	err := db.Offset(page.Normalize().Offset()).Limit(page.PageSize).Find(&roles).Error
	if err != nil {
		return nil, 0, domain.ErrInternal.ReWrap(err)
	}

	return roles, total, nil
}

func (r *rbacRepository) GetRoleByID(
	ctx context.Context,
	roleID uint64,
) (*entity.Role, error) {
	var role entity.Role
	err := r.db.WithContext(ctx).First(&role, roleID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound.ReWrap(err)
		}
		return nil, domain.ErrInternal.ReWrap(err)
	}
	return &role, nil
}

func (r *rbacRepository) CreateRoleWithPermissions(
	ctx context.Context,
	role *entity.Role,
	permissionIDs []uint64,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(permissionIDs) == 0 {
			return domain.ErrInternal.New()
		}

		if err := tx.Create(role).Error; err != nil {
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
				return domain.ErrConflict.ReWrap(err)
			}
			return domain.ErrInternal.ReWrap(err)
		}

		var rolePermissions []entity.RolePermission
		for _, pid := range permissionIDs {
			rolePermissions = append(rolePermissions, entity.RolePermission{
				RoleID:       role.ID,
				PermissionID: pid,
			})
		}

		if err := tx.Create(&rolePermissions).Error; err != nil {
			return domain.ErrInternal.ReWrap(err)
		}

		return nil
	})
}

func (r *rbacRepository) UpdateRoleWithPermissions(
	ctx context.Context,
	param *domain_repo.UpdateRoleParam,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// update role
		updates := make(map[string]any)
		if param.Name != nil {
			updates["name"] = *param.Name
		}

		if len(updates) > 0 {
			err := tx.Model(&entity.Role{}).
				Where("id = ?", param.ID).
				Updates(updates).Error
			if err != nil {
				var mysqlErr *mysql.MySQLError
				if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
					return domain.ErrConflict.ReWrap(err)
				}
				return domain.ErrInternal.ReWrap(err)
			}
		}

		// update permissions
		if param.PermissionIDs != nil {
			// 刪除舊關聯
			if err := tx.Where("role_id = ?", param.ID).Delete(&entity.RolePermission{}).Error; err != nil {
				return domain.ErrInternal.ReWrap(err)
			}

			// 建立新關聯
			permissionIDs := *param.PermissionIDs
			if len(permissionIDs) > 0 {
				var rolePermissions []entity.RolePermission
				for _, pid := range permissionIDs {
					rolePermissions = append(rolePermissions, entity.RolePermission{
						RoleID:       param.ID,
						PermissionID: pid,
					})
				}
				if err := tx.Create(&rolePermissions).Error; err != nil {
					return domain.ErrInternal.ReWrap(err)
				}
			}
		}

		return nil
	})
}

func (r *rbacRepository) GetRoleWithPermissions(
	ctx context.Context,
	roleID uint64,
) (*entity.RoleWithPermissions, error) {
	var rolePerms entity.RoleWithPermissions
	err := r.db.WithContext(ctx).
		Table(entity.Role{}.TableName()).
		Preload("Permissions").
		First(&rolePerms, roleID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound.ReWrap(err)
		}
		return nil, domain.ErrInternal.ReWrap(err)
	}

	return &rolePerms, nil
}

func (r *rbacRepository) ListPermissions(
	ctx context.Context,
	page domain.PageReq,
) ([]entity.Permission, int64, error) {
	var permissions []entity.Permission
	var total int64

	db := r.db.WithContext(ctx)

	// 取得總數
	if err := db.Model(&entity.Permission{}).Count(&total).Error; err != nil {
		return nil, 0, domain.ErrInternal.ReWrap(err)
	}

	// 分頁查詢
	err := db.Offset(page.Normalize().Offset()).Limit(page.PageSize).Find(&permissions).Error
	if err != nil {
		return nil, 0, domain.ErrInternal.ReWrap(err)
	}
	return permissions, total, nil
}

func (r *rbacRepository) GetRolePermissionIDs(
	ctx context.Context,
	roleID uint64,
) ([]uint64, error) {
	var ids []uint64
	err := r.db.WithContext(ctx).Model(&entity.RolePermission{}).
		Where("role_id = ?", roleID).
		Pluck("permission_id", &ids).Error
	if err != nil {
		return nil, domain.ErrInternal.ReWrap(err)
	}
	return ids, nil
}

func (r *rbacRepository) GetUserVisibleMenus(
	ctx context.Context,
	userID uint64,
) ([]entity.Permission, error) {
	// 目前用戶只能有一個 role
	var roleID uint64
	err := r.db.WithContext(ctx).Model(&entity.UserRole{}).
		Select("role_id").
		Where("user_id = ?", userID).
		Scan(&roleID).Error
	if err != nil {
		return nil, domain.ErrInternal.ReWrap(err)
	}
	if roleID == 0 {
		return []entity.Permission{}, nil
	}

	var permissions []entity.Permission
	sql := r.db.WithContext(ctx).Table("permission p").
		Joins("JOIN role_permission rp ON rp.permission_id = p.id").
		Where("p.type IN ?", []entity.PermissionType{entity.PermissionTypeDirectory, entity.PermissionTypeMenu})

	// 預設 admin role 時, 全部 menu 可見
	// TODO: role 表用 code 控制業務邏輯, 但後台也要讓用戶配置, 須溝通
	if roleID != 1 {
		sql = sql.Joins("JOIN role_permission rp ON rp.permission_id = p.id").
			Where("rp.role_id = ?", roleID).
			Where("p.type IN ?", []entity.PermissionType{entity.PermissionTypeDirectory, entity.PermissionTypeMenu})
	} else {
		sql = sql.Where("p.type IN ?", []entity.PermissionType{entity.PermissionTypeDirectory, entity.PermissionTypeMenu})
	}

	err = sql.Find(&permissions).Error
	if err != nil {
		return nil, domain.ErrInternal.ReWrap(err)
	}

	return permissions, nil
}

func (r *rbacRepository) GetRoleIDByUserID(ctx context.Context, userID uint64) (uint64, error) {
	var userRole entity.UserRole
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&userRole).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, domain.ErrNotFound.New()
		}
		return 0, domain.ErrInternal.ReWrap(err)
	}
	return userRole.RoleID, nil
}
