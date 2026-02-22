package domain

import "strconv"

// Status code range: 100000 - 199999
type Code int

const (
	// Common
	CodeSuccess             Code = 100000
	CodeInternalError       Code = 100001
	CodeParamInvalid        Code = 100002
	CodeNotFound            Code = 100003
	CodeUnauthorized        Code = 100004
	CodeForbidden           Code = 100005
	CodeConflict            Code = 100006
	CodeResourceExhausted   Code = 100007
	CodeUnprocessableEntity Code = 100008
	CodeTooManyRequests     Code = 100009

	// Custom
)

func (c Code) Int() int {
	return int(c)
}

func (c Code) String() string {
	return strconv.Itoa(int(c))
}
