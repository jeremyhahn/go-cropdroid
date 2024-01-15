package config

type Permission struct {
	OrganizationID uint64 `gorm:primaryKey" yaml:"orgId" json:"orgId"`
	FarmID         uint64 `gorm:primaryKey" yaml:"farmId" json:"farmId"`
	UserID         uint64 `gorm:primaryKey" yaml:"userId" json:"userId"`
	RoleID         uint64 `gorm:primaryKey" yaml:"roleId" json:"roleId"`
}

func NewPermission() *Permission {
	return &Permission{}
}

func CreatePermission(orgID, farmID, userID, roleID uint64) *Permission {
	return &Permission{
		OrganizationID: orgID,
		FarmID:         farmID,
		UserID:         userID,
		RoleID:         roleID}
}

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
