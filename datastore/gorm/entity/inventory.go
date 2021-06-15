package entity

import "time"

type InventoryEntity interface {
	GetID() int
	GetTypeID() int
}

type Inventory struct {
	ID              int        `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	InventoryTypeID int        `gorm:"foreign_key" json:"typeId"`
	DeviceID        int        `gorm:"foreign_key" json:"deviceId"`
	LifeExpectancy  int        `json:"lifeExpectancy"`
	StartDate       time.Time  `json:"startDate"`
	LastServiced    *time.Time `json:"lastServiced"`
	InventoryEntity `json:"-"`
}

func NewInventory() InventoryEntity {
	return &Inventory{}
}

func (inventory *Inventory) GetID() int {
	return inventory.ID
}

func (inventory *Inventory) GetInventoryTypeID() int {
	return inventory.InventoryTypeID
}

func (inventory *Inventory) GetDeviceID() int {
	return inventory.DeviceID
}

func (inventory *Inventory) GetLifeExpectancy() int {
	return inventory.LifeExpectancy
}

func (inventory *Inventory) GetStartDate() time.Time {
	return inventory.StartDate
}

func (inventory *Inventory) GetLastServiced() *time.Time {
	return inventory.LastServiced
}
