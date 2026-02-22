package repository

import (
	"context"

	"github.com/ichzzy/go_o11y_base/internal/domain/entity"
	domain_repo "github.com/ichzzy/go_o11y_base/internal/domain/repository"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

type UserRepoParam struct {
	dig.In

	DB *gorm.DB `name:"mysql"`
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(param UserRepoParam) domain_repo.UserRepository {
	return &userRepository{db: param.DB}
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, uint64, error) {
	var result struct {
		entity.User
		RoleID uint64 `gorm:"column:role_id"`
	}
	if err := r.db.WithContext(ctx).
		Table(entity.User{}.TableName()).
		Select("user.*, user_role.role_id").
		Joins("left join user_role on user.id = user_role.user_id").
		Where("user.email = ?", email).
		First(&result).Error; err != nil {
		return nil, 0, err
	}
	return &result.User, result.RoleID, nil
}

func (r *userRepository) CreateUser(ctx context.Context, user entity.User, roleID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 建立用戶
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		// 2. 建立用戶角色關係
		userRole := entity.UserRole{
			UserID: user.ID,
			RoleID: roleID,
		}
		if err := tx.Create(&userRole).Error; err != nil {
			return err
		}

		return nil
	})
}
