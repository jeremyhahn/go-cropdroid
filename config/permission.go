package config

type Permission struct {
	ID             uint64 `gorm:"primary_key:auto_increment"`
	OrganizationID uint64 `yaml:"orgId" json:"orgId"`
	FarmID         uint64 `yaml:"farmId" json:"farmId"`
	UserID         uint64 `yaml:"userId" json:"userId"`
	RoleID         uint64 `yaml:"roleId" json:"roleId"`
	//RoleID         int `gorm:"primary_key;auto_increment:false" yaml:"roleId" json:"roleId"`
}

func NewPermission(orgID, farmID, userID, roleID uint64) *Permission {
	return &Permission{
		OrganizationID: orgID,
		FarmID:         farmID,
		UserID:         userID,
		RoleID:         roleID}
}

func (perms *Permission) GetID() uint64 {
	return perms.ID
}

func (perms *Permission) GetOrgID() uint64 {
	return perms.OrganizationID
}

func (perms *Permission) GetFarmID() uint64 {
	return perms.FarmID
}

func (perms *Permission) GetUserID() uint64 {
	return perms.UserID
}

func (perms *Permission) GetRoleID() uint64 {
	return perms.RoleID
}
