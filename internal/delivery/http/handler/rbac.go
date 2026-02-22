package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ichzzy/go_o11y_base/internal/app/cx"
	"github.com/ichzzy/go_o11y_base/internal/delivery/http/payload"
	"github.com/ichzzy/go_o11y_base/internal/delivery/http/response"
	"github.com/ichzzy/go_o11y_base/internal/domain"
	domain_uc "github.com/ichzzy/go_o11y_base/internal/domain/usecase"
	"go.uber.org/dig"
)

type RbacHandlerParam struct {
	dig.In

	RbacUsecase domain_uc.RbacUsecase
}

type RbacHandler struct {
	rbacUsecase domain_uc.RbacUsecase
}

func NewRbacHandler(param RbacHandlerParam) *RbacHandler {
	return &RbacHandler{
		rbacUsecase: param.RbacUsecase,
	}
}

func (h *RbacHandler) ListRoles(c *gin.Context) {
	var req payload.ListRolesReq
	if err := c.ShouldBindQuery(&req); err != nil {
		_ = c.Error(domain.ErrParamInvalid.ReWrap(err))
		return
	}

	roles, total, err := h.rbacUsecase.ListRoles(c.Request.Context(), req.PageReq)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(domain.NewPageResp(roles, total, req.PageReq)))
}

func (h *RbacHandler) CreateRole(c *gin.Context) {
	var req payload.CreateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(domain.ErrParamInvalid.ReWrap(err))
		return
	}

	if err := h.rbacUsecase.CreateRoleWithPermissions(c.Request.Context(), req.Name, req.PermissionIDs); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(struct{}{}))
}

func (h *RbacHandler) UpdateRole(c *gin.Context) {
	var req payload.UpdateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(domain.ErrParamInvalid.ReWrap(err))
		return
	}

	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil {
		_ = c.Error(domain.ErrParamInvalid.ReWrap(err))
		return
	}

	if err := h.rbacUsecase.UpdateRoleWithPermissions(c.Request.Context(), roleID, req.Name, req.PermissionIDs); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(struct{}{}))
}

func (h *RbacHandler) ListPermissions(c *gin.Context) {
	var req payload.ListPermissionsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		_ = c.Error(domain.ErrParamInvalid.ReWrap(err))
		return
	}

	perms, total, err := h.rbacUsecase.ListPermissions(c.Request.Context(), req.PageReq)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(domain.NewPageResp(perms, total, req.PageReq)))
}

func (h *RbacHandler) ListRolePermissions(c *gin.Context) {
	roleIDStr := c.Param("roleID")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil {
		_ = c.Error(domain.ErrParamInvalid.ReWrap(errors.New("invalid role ID")))
		return
	}

	resp, err := h.rbacUsecase.GetRoleWithPermissions(c.Request.Context(), roleID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(resp.Permissions))
}

func (h *RbacHandler) ListMyVisibleMenus(c *gin.Context) {
	identity := cx.GetUserIdentity(c.Request.Context())

	perms, err := h.rbacUsecase.GetUserVisibleMenus(c.Request.Context(), identity.UserID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(perms))
}
