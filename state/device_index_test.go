package state

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/stretchr/testify/assert"
)

func TestDeviceIndex(t *testing.T) {

	deviceConfig := &config.DeviceStruct{
		ID:          1,
		FarmID:      uint64(1),
		Type:        "Test",
		Interval:    60,
		Description: "This is a device used for testing",
		Enable:      true,
		Notify:      false,
		URI:         "http://localhost"}

	store := NewDeviceIndex(1)
	store.Put(1, deviceConfig)

	device, ok := store.Get(1)
	assert.Equal(t, true, ok)
	assert.Equal(t, device, deviceConfig)
	assert.Equal(t, uint64(1), device.Identifier())
	assert.Equal(t, uint64(1), device.GetFarmID())
	assert.Equal(t, "Test", device.GetType())
}

func BenchmarkDeviceIndex(b *testing.B) {

	deviceConfig := &config.DeviceStruct{
		ID:          1,
		FarmID:      1,
		Type:        "Test",
		Interval:    60,
		Description: "This is a device used for testing",
		Enable:      true,
		Notify:      false,
		URI:         "http://localhost"}

	store := NewDeviceIndex(b.N)

	for n := 0; n < b.N; n++ {
		store.Put(uint64(n), deviceConfig)
	}
}
