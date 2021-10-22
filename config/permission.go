package config

type Permission struct {
	//ID               uint64 `gorm:"primaryKey:autoIncrement"`
	OrganizationID   uint64 `yaml:"orgId" json:"orgId"`
	FarmID           uint64 `yaml:"farmId" json:"farmId"`
	UserID           uint64 `yaml:"userId" json:"userId"`
	RoleID           uint64 `yaml:"roleId" json:"roleId"`
	PermissionConfig `yaml:"-" json:"="`
}

func NewPermission() PermissionConfig {
	return &Permission{}
}

func CreatePermission(orgID, farmID, userID, roleID uint64) PermissionConfig {
	return &Permission{
		OrganizationID: orgID,
		FarmID:         farmID,
		UserID:         userID,
		RoleID:         roleID}
}

// func (perms *Permission) GetID() uint64 {
// 	return perms.ID
// }

// func (perms *Permission) SetID(id uint64) {
// 	perms.ID = id
// }

func (perms *Permission) GetOrgID() uint64 {
	return perms.OrganizationID
}

func (perms *Permission) SetOrgID(id uint64) {
	perms.OrganizationID = id
}

func (perms *Permission) GetFarmID() uint64 {
	return perms.FarmID
}

func (perms *Permission) SetFarmID(id uint64) {
	perms.FarmID = id
}

func (perms *Permission) GetUserID() uint64 {
	return perms.UserID
}

func (perms *Permission) SetUserID(id uint64) {
	perms.UserID = id
}

func (perms *Permission) GetRoleID() uint64 {
	return perms.RoleID
}

func (perms *Permission) SetRoleID(id uint64) {
	perms.RoleID = id
}
