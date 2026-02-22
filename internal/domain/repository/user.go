package repository

import (
	"context"

	"github.com/ichzzy/go_o11y_base/internal/domain/entity"
)

type UserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*entity.User, uint64, error)
	CreateUser(ctx context.Context, user entity.User, roleID uint64) error
}
