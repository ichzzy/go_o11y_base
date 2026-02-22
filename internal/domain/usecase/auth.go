package usecase

import (
	"context"

	"github.com/ichzzy/go_o11y_base/internal/domain/entity"
)

type AuthUsecase interface {
	SignToken(ctx context.Context, userID, roleID uint64) (string, error)
	GenerateRefreshToken(ctx context.Context, userID, roleID uint64) (string, error)
	ValidateRefreshToken(ctx context.Context, tokenString string) (uint64, uint64, error)
	Refresh(ctx context.Context, oldRefreshToken string) (string, string, error)
	RevokeRefreshToken(ctx context.Context, tokenString string) error
	ParseToken(tokenString string) (*entity.Claims, error)
	Login(ctx context.Context, email, password string) (string, string, error)
	CreateUser(ctx context.Context, user entity.User, roleID uint64) error
}
