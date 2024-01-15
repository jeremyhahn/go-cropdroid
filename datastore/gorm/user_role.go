package gorm

// Junction table to map config.User.Roles
type UserRole struct {
	UserID uint64
	RoleID uint64
}
