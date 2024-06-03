package config

type Permission interface {
	GetOrgID() uint64
	SetOrgID(id uint64)
	GetFarmID() uint64
	SetFarmID(id uint64)
	GetUserID() uint64
	SetUserID(id uint64)
	GetRoleID() uint64
	SetRoleID(id uint64)
}

type PermissionStruct struct {
	OrganizationID uint64 `gorm:"primaryKey;autoIncrement=false" yaml:"orgId" json:"orgId"`
	FarmID         uint64 `gorm:"primaryKey;autoIncrement=false" yaml:"farmId" json:"farmId"`
	UserID         uint64 `gorm:"primaryKey;autoIncrement=false" yaml:"userId" json:"userId"`
	RoleID         uint64 `gorm:"primaryKey;autoIncrement=false" yaml:"roleId" json:"roleId"`
	Permission     `sql:"-" gorm:"-"`
}

func NewPermission() *PermissionStruct {
	return &PermissionStruct{}
}

func CreatePermissionStruct(orgID, farmID, userID, roleID uint64) *PermissionStruct {
	return &PermissionStruct{
		OrganizationID: orgID,
		FarmID:         farmID,
		UserID:         userID,
		RoleID:         roleID}
}

func (perms *PermissionStruct) TableName() string {
	return "permissions"
}

func (perms *PermissionStruct) GetOrgID() uint64 {
	return perms.OrganizationID
}

func (perms *PermissionStruct) SetOrgID(id uint64) {
	perms.OrganizationID = id
}

func (perms *PermissionStruct) GetFarmID() uint64 {
	return perms.FarmID
}

func (perms *PermissionStruct) SetFarmID(id uint64) {
	perms.FarmID = id
}

func (perms *PermissionStruct) GetUserID() uint64 {
	return perms.UserID
}

func (perms *PermissionStruct) SetUserID(id uint64) {
	perms.UserID = id
}

func (perms *PermissionStruct) GetRoleID() uint64 {
	return perms.RoleID
}

func (perms *PermissionStruct) SetRoleID(id uint64) {
	perms.RoleID = id
}
