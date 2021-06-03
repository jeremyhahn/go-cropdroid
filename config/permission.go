package config

type Permission struct {
	ID             int `gorm:"primary_key:auto_increment"`
	OrganizationID int `yaml:"orgId" json:"orgId"`
	FarmID         int `yaml:"farmId" json:"farmId"`
	UserID         int `yaml:"userId" json:"userId"`
	RoleID         int `yaml:"roleId" json:"roleId"`
	//RoleID         int `gorm:"primary_key;auto_increment:false" yaml:"roleId" json:"roleId"`
}

func NewPermission(orgID, farmID, userID, roleID int) *Permission {
	return &Permission{
		OrganizationID: orgID,
		FarmID:         farmID,
		UserID:         userID,
		RoleID:         roleID}
}

func (perms *Permission) GetID() int {
	return perms.ID
}

func (perms *Permission) GetOrgID() int {
	return perms.OrganizationID
}

func (perms *Permission) GetFarmID() int {
	return perms.FarmID
}

func (perms *Permission) GetUserID() int {
	return perms.UserID
}

func (perms *Permission) GetRoleID() int {
	return perms.RoleID
}
