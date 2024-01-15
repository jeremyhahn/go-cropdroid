package gorm

// Junction table to map config.Farm.Users
type UserFarm struct {
	UserID uint64
	FarmID uint64
}
