// +build integration

package test

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/controller"
	"github.com/stretchr/testify/assert"
)

func TestRoomClient(t *testing.T) {

	ctx := NewUnitTestContext()

	client := controller.NewRoomController(ctx, "http://room2.westland.dr", nil)
	entity, err := client.RoomStatus()

	assert.Nil(t, err)
	assert.True(t, entity.GetTempF0() > 0)

	ctx.GetLogger().Debugf("%+v\n", entity.GetTempF0())
}
