package usecase

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ichzzy/go_o11y_base/internal/app/cx"
	"github.com/ichzzy/go_o11y_base/internal/config"
	"github.com/ichzzy/go_o11y_base/internal/domain"
	"github.com/ichzzy/go_o11y_base/internal/domain/entity"
	"github.com/ichzzy/go_o11y_base/internal/domain/repository"
	domain_uc "github.com/ichzzy/go_o11y_base/internal/domain/usecase"
	"github.com/ichzzy/go_o11y_base/internal/lib/logs"
	"github.com/ichzzy/go_o11y_base/internal/utils/cryptos"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/dig"
	"golang.org/x/crypto/bcrypt"
)

var tracer = otel.Tracer("internal/usecase/auth")

type AuthUsecaseParam struct {
	dig.In

	Config   *config.Config `name:"config"`
	Redis    repository.RedisClientRepository
	UserRepo repository.UserRepository
}

type authUsecase struct {
	config   *config.Config
	redis    repository.RedisClientRepository
	userRepo repository.UserRepository
}

func NewAuthUsecase(param AuthUsecaseParam) domain_uc.AuthUsecase {
	return &authUsecase{
		config:   param.Config,
		redis:    param.Redis,
		userRepo: param.UserRepo,
	}
}

func (uc *authUsecase) SignToken(
	ctx context.Context,
	userID uint64,
	roleID uint64,
) (string, error) {
	_, span := tracer.Start(ctx, "AuthUsecase.SignToken")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("user_id", int64(userID)),
		attribute.Int64("role_id", int64(roleID)),
	)
	now := time.Now()
	claims := entity.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(uc.config.JWT.Duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    uc.config.App.ServiceName,
			Subject:   strconv.FormatUint(userID, 10),
		},
		RoleID: roleID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)

	// 使用 Ed25519 Private Key 簽名
	signedToken, err := token.SignedString(uc.config.JWT.PrivateKey)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "signed token failed")
		return "", domain.ErrInternal.ReWrap(err)
	}

	return signedToken, nil
}

func (uc *authUsecase) GenerateRefreshToken(
	ctx context.Context,
	userID uint64,
	roleID uint64,
) (string, error) {
	token, err := cryptos.GenerateToken(24)
	if err != nil {
		return "", domain.ErrInternal.ReWrap(err)
	}

	if err := uc.redis.Set(ctx,
		domain.RedisKeyRefreshToken.Format(token),
		fmt.Sprintf("%d:%d", userID, roleID),
		uc.config.JWT.RefreshDuration,
	); err != nil {
		return "", domain.ErrInternal.ReWrap(err)
	}

	if err := uc.redis.Set(ctx,
		domain.RedisKeyRefreshSession.Format(userID),
		token,
		uc.config.JWT.RefreshDuration,
	); err != nil {
		return "", domain.ErrInternal.ReWrap(err)
	}

	return token, nil
}

func (uc *authUsecase) ValidateRefreshToken(
	ctx context.Context,
	tokenString string,
) (uint64, uint64, error) {
	val, err := uc.redis.Get(ctx, domain.RedisKeyRefreshToken.Format(tokenString))
	if err != nil {
		return 0, 0, domain.ErrUnauthorized.New()
	}

	var userID, roleID uint64
	_, err = fmt.Sscanf(val, "%d:%d", &userID, &roleID)
	if err != nil {
		return 0, 0, domain.ErrInternal.ReWrap(err)
	}

	latestToken, err := uc.redis.Get(ctx, domain.RedisKeyRefreshSession.Format(userID))
	if err != nil {
		// Session 丟失或已過期，視為無效
		_ = uc.RevokeRefreshToken(ctx, tokenString)
		return 0, 0, domain.ErrUnauthorized.New()
	}

	if latestToken != tokenString {
		cx.GetLogger(ctx).WithFields(logs.Fields{
			"userID": userID,
			"roleID": roleID,
		}).Warn("refresh token reuse detected")
		_ = uc.RevokeRefreshToken(ctx, tokenString)
		return 0, 0, domain.ErrUnauthorized.New()
	}

	return userID, roleID, nil
}

func (uc *authUsecase) Refresh(
	ctx context.Context,
	oldRefreshToken string,
) (string, string, error) {
	// 1. 驗證 Token
	userID, roleID, err := uc.ValidateRefreshToken(ctx, oldRefreshToken)
	if err != nil {
		return "", "", err
	}

	// 2. Rotate token
	newAccessToken, err := uc.SignToken(ctx, userID, roleID)
	if err != nil {
		return "", "", err
	}
	newRefreshToken, err := uc.GenerateRefreshToken(ctx, userID, roleID)
	if err != nil {
		return "", "", err
	}

	// 3. 移除舊 refresh token
	_ = uc.redis.Del(ctx, domain.RedisKeyRefreshToken.Format(oldRefreshToken))

	return newAccessToken, newRefreshToken, nil
}

func (uc *authUsecase) RevokeRefreshToken(
	ctx context.Context,
	tokenString string,
) error {
	return uc.redis.Del(ctx, domain.RedisKeyRefreshToken.Format(tokenString))
}

func (uc *authUsecase) ParseToken(
	tokenString string,
) (*entity.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &entity.Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return uc.config.JWT.PublicKey, nil
	})
	if err != nil {
		return nil, domain.ErrInternal.ReWrap(err)
	}
	if !token.Valid {
		return nil, domain.ErrUnauthorized.New()
	}

	if claims, ok := token.Claims.(*entity.Claims); ok {
		return claims, nil
	}

	return nil, domain.ErrUnprocessableEntity.New()
}

func (uc *authUsecase) Login(
	ctx context.Context,
	email string,
	password string,
) (string, string, error) {
	user, roleID, err := uc.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		cx.GetLogger(ctx).WithError(err).Warn("GetUserByEmail failed")
		return "", "", domain.ErrUnauthorized.New()
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		cx.GetLogger(ctx).WithError(err).Warn("password mismatch")
		return "", "", domain.ErrUnauthorized.New()
	}

	accessToken, err := uc.SignToken(ctx, user.ID, roleID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := uc.GenerateRefreshToken(ctx, user.ID, roleID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (uc *authUsecase) CreateUser(
	ctx context.Context,
	user entity.User,
	roleID uint64,
) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.ErrInternal.ReWrap(err)
	}
	user.Password = string(hash)
	user.Status = entity.UserStatusActive

	if err := uc.userRepo.CreateUser(ctx, user, roleID); err != nil {
		return domain.ErrInternal.ReWrap(err)
	}
	return nil
}
