package entity

type User struct {
	ID        uint64     `gorm:"primaryKey" json:"id,string"`
	Email     string     `json:"email"`
	Password  string     `json:"-"`
	Status    UserStatus `json:"status"`
	CreatedAt int64      `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64      `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (User) TableName() string {
	return "user"
}

type UserStatus uint8

const (
	UserStatusUndefined UserStatus = 0
	UserStatusActive    UserStatus = 1
	UserStatusInactive  UserStatus = 2
)

type UserIdentity struct {
	UserID uint64
	RoleID uint64
}
