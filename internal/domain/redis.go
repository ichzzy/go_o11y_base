package domain

import "fmt"

type RedisKey string

func (rk RedisKey) Format(args ...any) string {
	return fmt.Sprintf(string(rk), args...)
}

const (
	RedisKeyPrefix         RedisKey = "go_o11y_base"
	RedisKeyRefreshToken   RedisKey = RedisKeyPrefix + ":refreshToken:%s"   // Refresh Token (儲存 UserID:RoleID)
	RedisKeyRefreshSession RedisKey = RedisKeyPrefix + ":refreshSession:%d" // UserID (儲存最新的 Refresh Token); PM 無規格, 目前 session 顆粒度為 user
)

type LuaScript = string

const (
	LuaScriptUnlockMutex LuaScript = `
	return redis.call("GET", KEYS[1]) == ARGV[1] 
		and redis.call("DEL", KEYS[1])
		or 0
	`
)
