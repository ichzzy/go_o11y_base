package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ichzzy/go_o11y_base/internal/delivery/http/payload"
	"github.com/ichzzy/go_o11y_base/internal/delivery/http/response"
	"github.com/ichzzy/go_o11y_base/internal/domain"
	"github.com/ichzzy/go_o11y_base/internal/domain/entity"
	domain_uc "github.com/ichzzy/go_o11y_base/internal/domain/usecase"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/dig"
)

var Tracer = otel.Tracer("internal/delivery/http/handler")

type AuthHandlerParam struct {
	dig.In

	AuthUseCase domain_uc.AuthUsecase
}

type AuthHandler struct {
	authUseCase domain_uc.AuthUsecase
}

func NewAuthHandler(param AuthHandlerParam) *AuthHandler {
	return &AuthHandler{
		authUseCase: param.AuthUseCase,
	}
}

func (h *AuthHandler) DevLogin(c *gin.Context) {
	ctx, span := Tracer.Start(c.Request.Context(), "AuthHandler.DevLogin")
	defer span.End()

	var req payload.DevLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "parameter binding failed")
		_ = c.Error(domain.ErrParamInvalid.ReWrap(err))
		return
	}

	span.SetAttributes(
		attribute.Int64("req.user_id", int64(req.UserID)),
		attribute.Int64("req.role_id", int64(req.RoleID)),
	)

	token, err := h.authUseCase.SignToken(ctx, req.UserID, req.RoleID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "sign token failed")
		_ = c.Error(err)
		return
	}

	refreshToken, err := h.authUseCase.GenerateRefreshToken(ctx, req.UserID, req.RoleID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "generate refresh token failed")
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(payload.DevLoginResp{
		AccessToken:  token,
		RefreshToken: refreshToken,
	}))
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req payload.RefreshTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(domain.ErrParamInvalid.ReWrap(err))
		return
	}

	accessToken, refreshToken, err := h.authUseCase.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(payload.RefreshTokenResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req payload.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(domain.ErrParamInvalid.ReWrap(err))
		return
	}

	token, refreshToken, err := h.authUseCase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(payload.LoginResp{
		AccessToken:  token,
		RefreshToken: refreshToken,
	}))
}

func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req payload.CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(domain.ErrParamInvalid.ReWrap(err))
		return
	}

	user := entity.User{
		Email:    req.Email,
		Password: req.Password,
	}

	err := h.authUseCase.CreateUser(c.Request.Context(), user, req.RoleID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(payload.CreateUserResp{}))
}
