package data

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/util"
)

func CreateTestOrganization1(idGenerator util.IdGenerator) *config.Organization {
	orgName := "Test Org 1"
	farm1Name := "Test Farm 1"
	farm2Name := "Test Farm 2"
	farm1 := config.NewFarm()
	farm1.SetID(idGenerator.NewStringID(farm1Name))
	farm1.SetName(farm1Name)
	farm1.SetMode("test")
	farm1.SetDevices([]*config.Device{
		{
			Type: "server",
			Settings: []*config.DeviceSetting{
				{
					Key:   "name",
					Value: farm1Name},
				{
					Key:   "mode",
					Value: "test"},
				{
					Key:   "interval",
					Value: "59"}}}})

	farm2 := config.NewFarm()
	farm2.SetID(idGenerator.NewStringID(farm2Name))
	farm2.SetMode(farm2Name)
	farm2.SetDevices([]*config.Device{
		{
			Type: "server",
			Settings: []*config.DeviceSetting{
				{
					Key:   "name",
					Value: farm2Name},
				{
					Key:   "mode",
					Value: "test2"},
				{
					Key:   "interval",
					Value: "60"}}}})

	org := config.NewOrganization()
	org.SetID(idGenerator.NewStringID(orgName))
	org.SetName(orgName)
	org.SetFarms([]*config.Farm{farm1, farm2})

	return org
}

func CreateTestOrganization2(idGenerator util.IdGenerator) *config.Organization {
	orgName := "Test Org 2"
	farm3Name := "Test Farm 3"
	farm4Name := "Test Farm 4"
	farm3 := config.NewFarm()
	farm3.SetID(idGenerator.NewStringID(farm3Name))
	farm3.SetName(farm3Name)
	farm3.SetMode("test")
	farm3.SetDevices([]*config.Device{
		{
			Type: "server",
			Settings: []*config.DeviceSetting{
				{
					Key:   "name",
					Value: farm3Name},
				{
					Key:   "mode",
					Value: "test"},
				{
					Key:   "interval",
					Value: "59"}}}})

	farm4 := config.NewFarm()
	farm4.SetID(idGenerator.NewStringID(farm4Name))
	farm4.SetName(farm4Name)
	farm4.SetMode("test")
	farm4.SetDevices([]*config.Device{
		{
			Type: "server",
			Settings: []*config.DeviceSetting{
				{
					Key:   "name",
					Value: farm4Name},
				{
					Key:   "mode",
					Value: "test2"},
				{
					Key:   "interval",
					Value: "60"}}}})

	org := config.NewOrganization()
	org.SetID(idGenerator.NewStringID(orgName))
	org.SetName(orgName)
	org.SetFarms([]*config.Farm{farm3, farm4})

	return org
}
