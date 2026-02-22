package domain

import "net/http"

var (
	// Common
	ErrInternal            = New(CodeInternalError, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	ErrParamInvalid        = New(CodeParamInvalid, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	ErrNotFound            = New(CodeNotFound, http.StatusNotFound, http.StatusText(http.StatusNotFound))
	ErrUnauthorized        = New(CodeUnauthorized, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	ErrForbidden           = New(CodeForbidden, http.StatusForbidden, http.StatusText(http.StatusForbidden))
	ErrConflict            = New(CodeConflict, http.StatusConflict, http.StatusText(http.StatusConflict))
	ErrResourceExhausted   = New(CodeResourceExhausted, http.StatusTooManyRequests, http.StatusText(http.StatusUnprocessableEntity))
	ErrUnprocessableEntity = New(CodeUnprocessableEntity, http.StatusUnprocessableEntity, http.StatusText(http.StatusUnprocessableEntity))
	ErrTooManyRequests     = New(CodeTooManyRequests, http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))

	// Custom
)
