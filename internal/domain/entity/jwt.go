package entity

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	jwt.RegisteredClaims
	RoleID uint64 `json:"role_id,string"`
}
