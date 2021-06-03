package entity

import "time"

type InventoryTypeEntity interface {
	GetID() int
	GetNname() string
	GetExpiration() *time.Time
}

type InventoryType struct {
	ID                  int    `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	Name                string `gorm:"size:255" json:"name"`
	Description         string `gorm:"size:255" json:"description"`
	Image               string `gorm:"size:255" json:"image"`
	LifeExpectancy      int    `json:"lifeExpectancy"`
	MaintenanceCycle    int    `json:"maintenanceCycle"`
	ProductPage         string `gorm:"size:255" json:"productPage"`
	InventoryTypeEntity `json:"-"`
}

func NewInventoryType() InventoryTypeEntity {
	return &InventoryType{}
}

func (inventoryType *InventoryType) GetID() int {
	return inventoryType.ID
}

func (inventoryType *InventoryType) GetName() string {
	return inventoryType.Name
}

func (inventoryType *InventoryType) GetDescription() string {
	return inventoryType.Description
}

func (inventoryType *InventoryType) GetImage() string {
	return inventoryType.Image
}

func (inventoryType *InventoryType) GetLifeExpectancy() int {
	return inventoryType.LifeExpectancy
}

func (inventoryType *InventoryType) GetMaintenanceCycle() int {
	return inventoryType.MaintenanceCycle
}

func (inventoryType *InventoryType) GetProductPage() string {
	return inventoryType.ProductPage
}
