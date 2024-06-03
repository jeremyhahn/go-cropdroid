package util

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
)

type IdSetter interface {
	SetIds(farm *config.FarmStruct) *config.FarmStruct
	SetDeviceIds(farmID uint64, devices []*config.DeviceStruct) []*config.DeviceStruct
	SetDeviceSettingIds(device *config.DeviceStruct, deviceSettings []*config.DeviceSettingStruct) []*config.DeviceSettingStruct
	SetMetricIds(deviceID uint64, metrics []*config.MetricStruct) []*config.MetricStruct
	SetChannelsIds(farmID, deviceID uint64, channels []*config.ChannelStruct) []*config.ChannelStruct
	SetConditionIds(channelID, deviceID uint64, conditions []*config.ConditionStruct) []*config.ConditionStruct
	SetScheduleIds(farmID, deviceID, channelID uint64, schedules []*config.ScheduleStruct) []*config.ScheduleStruct
	SetUserIds(users []*config.UserStruct) []*config.UserStruct
	SetRoleIds(roles []*config.RoleStruct) []*config.RoleStruct
	SetWorkflowIds(farmID uint64, workflows []*config.WorkflowStruct) []*config.WorkflowStruct
	SetWorkflowStepIds(workflowID uint64, workflowSteps []*config.WorkflowStepStruct) []*config.WorkflowStepStruct
	SetCustomerIds(customer *config.CustomerStruct) *config.CustomerStruct
}

type KeyValueSetter struct {
	idGenerator IdGenerator
	IdSetter
}

func NewIdSetter(idGenerator IdGenerator) IdSetter {
	return &KeyValueSetter{idGenerator: idGenerator}
}

func (setter *KeyValueSetter) SetIds(farm *config.FarmStruct) *config.FarmStruct {
	farmID := farm.ID
	if farmID == 0 {
		farmID := setter.idGenerator.NewFarmID(farm.GetOrganizationID(), farm.GetName())
		farm.SetID(farmID)
	}
	setter.SetDeviceIds(farmID, farm.GetDevices())
	setter.SetUserIds(farm.GetUsers())
	setter.SetWorkflowIds(farmID, farm.GetWorkflows())
	return farm
}

func (setter *KeyValueSetter) SetDeviceIds(farmID uint64, devices []*config.DeviceStruct) []*config.DeviceStruct {
	for _, device := range devices {
		if device.ID == 0 {
			deviceID := setter.idGenerator.NewDeviceID(farmID, device.GetType())
			device.SetID(deviceID)
		}
		if device.GetFarmID() == 0 {
			device.SetFarmID(farmID)
		}
		setter.SetDeviceSettingIds(device, device.GetSettings())
	}
	return devices
}

func (setter *KeyValueSetter) SetDeviceSettingIds(device *config.DeviceStruct,
	deviceSettings []*config.DeviceSettingStruct) []*config.DeviceSettingStruct {

	deviceID := device.ID
	for _, deviceSetting := range deviceSettings {
		if deviceSetting.ID == 0 {
			deviceSettingID := setter.idGenerator.NewDeviceSettingID(deviceID, deviceSetting.GetKey())
			deviceSetting.SetID(deviceSettingID)
		}
		setter.SetMetricIds(deviceID, device.GetMetrics())
		setter.SetChannelsIds(device.GetFarmID(), deviceID, device.GetChannels())
	}
	return deviceSettings
}

func (setter *KeyValueSetter) SetMetricIds(deviceID uint64, metrics []*config.MetricStruct) []*config.MetricStruct {
	for _, metric := range metrics {
		if metric.ID == 0 {
			metricID := setter.idGenerator.NewMetricID(deviceID, metric.GetKey())
			metric.SetID(metricID)
		}
		if metric.GetDeviceID() == 0 {
			metric.SetDeviceID(deviceID)
		}
	}
	return metrics
}

func (setter *KeyValueSetter) SetChannelsIds(farmID, deviceID uint64, channels []*config.ChannelStruct) []*config.ChannelStruct {
	for _, channel := range channels {
		if channel.ID == 0 {
			channelID := setter.idGenerator.NewChannelID(deviceID, channel.GetName())
			channel.SetID(channelID)
		}
		if channel.GetDeviceID() == 0 {
			channel.SetDeviceID(deviceID)
		}
		setter.SetConditionIds(channel.ID, deviceID, channel.GetConditions())
		setter.SetScheduleIds(farmID, deviceID, channel.ID, channel.GetSchedule())
	}
	return channels
}

func (setter *KeyValueSetter) SetConditionIds(channelID, deviceID uint64, conditions []*config.ConditionStruct) []*config.ConditionStruct {
	for _, condition := range conditions {
		if condition.ID == 0 {
			conditionKey := fmt.Sprintf("%d-%d-%d-%d-%s-%2f", deviceID,
				condition.GetWorkflowID(), condition.GetChannelID(),
				condition.GetMetricID(), condition.GetComparator(),
				condition.GetThreshold())
			conditionID := setter.idGenerator.NewConditionID(deviceID, conditionKey)
			condition.SetID(conditionID)
		}
		if condition.GetChannelID() == 0 {
			condition.SetChannelID(channelID)
		}
	}
	return conditions
}

func (setter *KeyValueSetter) SetScheduleIds(farmID, deviceID, channelID uint64, schedules []*config.ScheduleStruct) []*config.ScheduleStruct {
	for _, schedule := range schedules {
		if schedule.ID == 0 {
			scheduleKey := fmt.Sprintf("%d-%d-%d-%s-%s-%d-%d", farmID, deviceID,
				schedule.GetChannelID(), schedule.GetStartDate(),
				schedule.GetEndDate(), schedule.GetFrequency(),
				schedule.GetCount())
			scheduleID := setter.idGenerator.NewScheduleID(deviceID, scheduleKey)
			schedule.SetID(scheduleID)
		}
		if schedule.GetChannelID() == 0 {
			schedule.SetChannelID(channelID)
		}
	}
	return schedules
}

func (setter *KeyValueSetter) SetUserIds(users []*config.UserStruct) []*config.UserStruct {
	for _, user := range users {
		if user.ID == 0 {
			user.SetID(setter.idGenerator.NewUserID(user.GetEmail()))
		}
		setter.SetRoleIds(user.GetRoles())
	}
	return users
}

func (setter *KeyValueSetter) SetRoleIds(roles []*config.RoleStruct) []*config.RoleStruct {
	for _, role := range roles {
		if role.ID == 0 {
			role.SetID(setter.idGenerator.NewRoleID(role.GetName()))
		}
	}
	return roles
}

func (setter *KeyValueSetter) SetWorkflowIds(farmID uint64, workflows []*config.WorkflowStruct) []*config.WorkflowStruct {
	for _, workflow := range workflows {
		if workflow.ID == 0 {
			workflowID := setter.idGenerator.NewWorkflowID(farmID, workflow.GetName())
			workflow.SetID(workflowID)
		}
		setter.SetWorkflowStepIds(workflow.ID, workflow.GetSteps())
	}
	return workflows
}

func (setter *KeyValueSetter) SetWorkflowStepIds(workflowID uint64, workflowSteps []*config.WorkflowStepStruct) []*config.WorkflowStepStruct {
	for _, workflowStep := range workflowSteps {
		if workflowStep.ID == 0 {
			workflowStepKey := fmt.Sprintf("%d-%d-%d-%d-%d", workflowID, workflowStep.GetDeviceID(),
				workflowStep.GetChannelID(), workflowStep.GetDuration(), workflowStep.GetState())
			workflowStepID := setter.idGenerator.NewWorkflowStepID(workflowID, workflowStepKey)
			workflowStep.SetID(workflowStepID)
		}
		if workflowStep.GetWorkflowID() == 0 {
			workflowStep.SetWorkflowID(workflowID)
		}
	}
	return workflowSteps
}

func (setter *KeyValueSetter) SetCustomerIds(customer *config.CustomerStruct) *config.CustomerStruct {
	if customer.ID == 0 {
		customer.ID = setter.idGenerator.NewCustomerID(customer.Email)
	}
	return customer
}
