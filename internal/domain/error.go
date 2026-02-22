package domain

import (
	"fmt"
	"runtime"
)

type AppError struct {
	code    Code   // 全局自定義狀態碼
	status  int    // HTTP Status Code
	message string // 訊息

	err        error
	stackTrace string
}

func New(code Code, status int, message string) AppError {
	return AppError{
		code:    code,
		status:  status,
		message: message,
	}
}

func (appErr AppError) New() AppError {
	appErr.stackTrace = stackTrace()
	return appErr
}

func (appErr AppError) Code() Code {
	return appErr.code
}

func (appErr AppError) Status() int {
	return appErr.status
}

func (appErr AppError) Message() string {
	return appErr.message
}

func (appErr AppError) Is(other error) bool {
	oe, ok := other.(AppError)
	return ok && oe.code == appErr.code
}

func (appErr AppError) Error() string {
	if appErr.err != nil {
		return appErr.err.Error()
	}

	return ""
}

func (appErr AppError) StackTrace() string {
	return appErr.stackTrace
}

// 這會揭露到 client 端, 使用前注意
func (appErr AppError) ReMsg(message string) AppError {
	appErr.message = message
	if appErr.stackTrace == "" {
		appErr.stackTrace = stackTrace()
	}

	return appErr
}

// 這會揭露到 client 端, 使用前注意
func (appErr AppError) ReMsgf(format string, args ...any) AppError {
	appErr.message = fmt.Sprintf(format, args...)
	if appErr.stackTrace == "" {
		appErr.stackTrace = stackTrace()
	}

	return appErr
}

// Error message 只會揭露在 log, 使用前可與 ReMsg 比較
// 會串聯 stack trace
func (appErr AppError) ReWrap(err error) AppError {
	appErr.err = err
	appErr.stackTrace = stackTrace()
	if wrappedAppErr, ok := err.(AppError); ok {
		appErr.stackTrace = appErr.stackTrace + ";" + wrappedAppErr.StackTrace()
	}
	return appErr
}

// Error message 只會揭露在 log, 使用前可與 ReMsg 比較
func (appErr AppError) ReWrapf(format string, args ...any) AppError {
	appErr.err = fmt.Errorf(format, args...)
	appErr.stackTrace = stackTrace()
	return appErr
}

func stackTrace() string {
	_, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s:%d", file, line)
}
