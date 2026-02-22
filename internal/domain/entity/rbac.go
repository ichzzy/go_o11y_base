package entity

// Role 角色
type Role struct {
	ID        uint64 `gorm:"primaryKey" json:"id,string"`
	Name      string `json:"name"`
	CreatedAt int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (Role) TableName() string {
	return "role"
}

// Permission 權限 (目錄、選單、API)
type Permission struct {
	ID         uint64         `gorm:"primaryKey" json:"id,string"`
	ParentID   uint64         `json:"parent_id,string"`
	Name       string         `json:"name"`
	Code       string         `json:"code"`
	Type       PermissionType `json:"type"`
	HTTPMethod string         `json:"http_method"`
	HTTPPath   string         `json:"http_path"`
	CreatedAt  int64          `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt  int64          `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (Permission) TableName() string {
	return "permission"
}

type PermissionType uint8

const (
	PermissionTypeUndefined PermissionType = 0
	PermissionTypeDirectory PermissionType = 1 // 目錄
	PermissionTypeMenu      PermissionType = 2 // 選單
	PermissionTypeAPI       PermissionType = 3 // API
)

// RolePermission 角色與權限關聯
type RolePermission struct {
	ID           uint64 `gorm:"primaryKey" json:"id,string"`
	RoleID       uint64 `json:"role_id,string"`
	PermissionID uint64 `json:"permission_id,string"`
	CreatedAt    int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt    int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (RolePermission) TableName() string {
	return "role_permission"
}

// UserRole 用戶與角色關聯
type UserRole struct {
	ID        uint64 `gorm:"primaryKey" json:"id,string"`
	UserID    uint64 `json:"user_id,string"`
	RoleID    uint64 `json:"role_id,string"`
	CreatedAt int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`

	Role *Role `gorm:"foreignKey:ID;references:RoleID" json:"-"`
}

func (UserRole) TableName() string {
	return "user_role"
}

type PolicyRule struct {
	RoleID uint64
	Path   string
	Method string
}

type RoleWithPermissions struct {
	Role        Role         `gorm:"embedded" json:"role"`
	Permissions []Permission `gorm:"many2many:role_permission;foreignKey:ID;joinForeignKey:RoleID;References:ID;joinReferences:PermissionID" json:"permissions"`
}
