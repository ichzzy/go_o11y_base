package payload

type DevLoginReq struct {
	UserID uint64 `json:"user_id"`
	RoleID uint64 `json:"role_id"`
}
type DevLoginResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenReq struct {
	RefreshToken string `json:"refresh_token"`
}
type RefreshTokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type CreateUserReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	RoleID   uint64 `json:"role_id"`
}
type CreateUserResp struct {
}
