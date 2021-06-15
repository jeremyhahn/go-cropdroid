// +build broken

package provisioner

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/stretchr/testify/assert"

	gormstore "github.com/jeremyhahn/go-cropdroid/datastore/gorm"
)

func TestProvisioner(t *testing.T) {

	app := NewIntegrationTest().app

	user := &model.User{
		Email:    "root@localhost",
		Password: "dev"}

	farmDAO := gormstore.NewFarmDAO(app.Logger, app.GORM)
	provisioner := NewGormFarmProvisioner(app.Logger, app.GORM, app.Location, farmDAO)

	farm, err := provisioner.Provision(user)
	assert.Nil(t, err)

	assert.NotNil(t, farm)
	assert.Equal(t, 4, len(farm.GetDevices()))

	farms, err := farmDAO.GetAll()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(farms))

	CurrentTest.Cleanup()
}

func TestProvisionerMultipleFarms(t *testing.T) {

	app := NewIntegrationTest().app

	user := &model.User{
		Email:    "root@localhost",
		Password: "dev"}

	farmDAO := gormstore.NewFarmDAO(app.Logger, app.GORM)
	provisioner := NewGormFarmProvisioner(app.Logger, app.GORM, app.Location, farmDAO)

	farm1, err := provisioner.Provision(user)
	assert.Nil(t, err)

	farm2, err := provisioner.Provision(user)
	assert.Nil(t, err)

	farm3, err := provisioner.Provision(user)
	assert.Nil(t, err)

	assert.NotNil(t, farm1)
	assert.Equal(t, 4, len(farm1.GetDevices()))
	assert.Equal(t, 4, len(farm2.GetDevices()))
	assert.Equal(t, 4, len(farm3.GetDevices()))

	farms, err := farmDAO.GetAll()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(farms))

	CurrentTest.Cleanup()
}
