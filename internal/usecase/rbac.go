package usecase

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/ichzzy/go_o11y_base/internal/app/cx"
	"github.com/ichzzy/go_o11y_base/internal/config"
	"github.com/ichzzy/go_o11y_base/internal/domain"
	"github.com/ichzzy/go_o11y_base/internal/domain/entity"
	domain_repo "github.com/ichzzy/go_o11y_base/internal/domain/repository"
	domain_uc "github.com/ichzzy/go_o11y_base/internal/domain/usecase"
	"github.com/ichzzy/go_o11y_base/internal/lib/logs"
	"go.uber.org/dig"
)

type RbacUsecaseParam struct {
	dig.In

	RbacRepo domain_repo.RbacRepository
	Config   *config.Config `name:"config"`
}

type rbacUsecase struct {
	enforcer *casbin.Enforcer
	rbacRepo domain_repo.RbacRepository
	config   *config.Config
	mu       sync.RWMutex
}

func NewRbacUsecase(param RbacUsecaseParam) (domain_uc.RbacUsecase, error) {
	modelPath := getModelPath()
	m, err := model.NewModelFromFile(modelPath)
	if err != nil {
		return nil, domain.ErrInternal.ReWrapf("load casbin model failed (path: %s): %w", modelPath, err)
	}

	e, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, domain.ErrInternal.ReWrapf("new enforcer failed: %w", err)
	}

	return &rbacUsecase{
		enforcer: e,
		rbacRepo: param.RbacRepo,
		config:   param.Config,
	}, nil
}

func getModelPath() string {
	defaultPath := "config/casbin_model.conf"
	// default workdir
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath
	}
	// 試著從上兩層目錄找 (單元測試時)
	if _, err := os.Stat("../../" + defaultPath); err == nil {
		return "../../" + defaultPath
	}
	return defaultPath
}

func (uc *rbacUsecase) LoadPolicy(ctx context.Context) error {
	// 1. 先從 Repo 查詢所有規則
	rules, err := uc.rbacRepo.GetPolicyRules(ctx)
	if err != nil {
		return domain.ErrInternal.ReWrap(err)
	}

	// 2. 寫鎖, 刷新記憶體rules
	uc.mu.Lock()
	defer uc.mu.Unlock()

	uc.enforcer.ClearPolicy()

	for _, rule := range rules {
		// p, role:{id}, {path}, {method}
		sub := fmt.Sprintf("role:%d", rule.RoleID)
		obj := rule.Path
		act := rule.Method

		if _, err := uc.enforcer.AddPolicy(sub, obj, act); err != nil {
			return domain.ErrInternal.ReWrap(err)
		}
	}

	return nil
}

func (uc *rbacUsecase) Enforce(sub string, obj string, act string) (bool, error) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	sub = "role:" + sub
	return uc.enforcer.Enforce(sub, obj, act)
}

func (uc *rbacUsecase) ListRoles(
	ctx context.Context,
	page domain.PageReq,
) ([]entity.Role, int64, error) {
	roles, total, err := uc.rbacRepo.ListRoles(ctx, page)
	if err != nil {
		return nil, 0, domain.ErrInternal.ReWrap(err)
	}
	return roles, total, nil
}

func (uc *rbacUsecase) CreateRoleWithPermissions(
	ctx context.Context,
	name string,
	permissionIDs []uint64,
) error {
	role := &entity.Role{
		Name: name,
	}
	if err := uc.rbacRepo.CreateRoleWithPermissions(ctx, role, permissionIDs); err != nil {
		return err
	}

	go func(logger logs.AppLogger) {
		if err := uc.LoadPolicy(ctx); err != nil {
			logger.Errorf("reload policy failed: %v", err)
		}
	}(cx.GetLogger(ctx))

	return nil
}

func (uc *rbacUsecase) UpdateRoleWithPermissions(
	ctx context.Context,
	id uint64,
	name string,
	permissionIDs []uint64,
) error {
	if err := uc.rbacRepo.UpdateRoleWithPermissions(ctx, &domain_repo.UpdateRoleParam{
		ID:            id,
		Name:          &name,
		PermissionIDs: &permissionIDs,
	}); err != nil {
		return domain.ErrInternal.ReWrap(err)
	}

	go func(logger logs.AppLogger) {
		if err := uc.LoadPolicy(ctx); err != nil {
			logger.Errorf("reload policy failed: %v", err)
		}
	}(cx.GetLogger(ctx))

	return nil
}

func (uc *rbacUsecase) GetRoleWithPermissions(
	ctx context.Context,
	roleID uint64,
) (*entity.RoleWithPermissions, error) {
	roleWithPerms, err := uc.rbacRepo.GetRoleWithPermissions(ctx, roleID)
	if err != nil {
		return nil, domain.ErrInternal.ReWrap(err)
	}
	return roleWithPerms, nil
}

func (uc *rbacUsecase) ListPermissions(
	ctx context.Context,
	page domain.PageReq,
) ([]entity.Permission, int64, error) {
	perms, total, err := uc.rbacRepo.ListPermissions(ctx, page)
	if err != nil {
		return nil, 0, domain.ErrInternal.ReWrap(err)
	}
	return perms, total, nil
}

func (uc *rbacUsecase) GetUserVisibleMenus(
	ctx context.Context,
	userID uint64,
) ([]entity.Permission, error) {
	perms, err := uc.rbacRepo.GetUserVisibleMenus(ctx, userID)
	if err != nil {
		return nil, domain.ErrInternal.ReWrap(err)
	}
	return perms, nil
}
