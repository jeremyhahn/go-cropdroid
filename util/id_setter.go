package util

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/config"
)

type IdSetter interface {
	SetIds(farm *config.Farm) *config.Farm
	SetDeviceIds(farmID uint64, devices []*config.Device) []*config.Device
	SetDeviceSettingIds(device *config.Device, deviceSettings []*config.DeviceSetting) []*config.DeviceSetting
	SetMetricIds(deviceID uint64, metrics []*config.Metric) []*config.Metric
	SetChannelsIds(farmID, deviceID uint64, channels []*config.Channel) []*config.Channel
	SetConditionIds(channelID, deviceID uint64, conditions []*config.Condition) []*config.Condition
	SetScheduleIds(farmID, deviceID, channelID uint64, schedules []*config.Schedule) []*config.Schedule
	SetUserIds(users []*config.User) []*config.User
	SetRoleIds(roles []*config.Role) []*config.Role
	SetWorkflowIds(farmID uint64, workflows []*config.Workflow) []*config.Workflow
	SetWorkflowStepIds(workflowID uint64, workflowSteps []*config.WorkflowStep) []*config.WorkflowStep
}

type KeyValueSetter struct {
	idGenerator IdGenerator
	IdSetter
}

func NewIdSetter(idGenerator IdGenerator) IdSetter {
	return &KeyValueSetter{idGenerator: idGenerator}
}

func (setter *KeyValueSetter) SetIds(farm *config.Farm) *config.Farm {
	farmID := farm.GetID()
	if farmID == 0 {
		farmID := setter.idGenerator.NewFarmID(farm.GetOrganizationID(), farm.GetName())
		farm.SetID(farmID)
	}
	setter.SetDeviceIds(farmID, farm.GetDevices())
	setter.SetUserIds(farm.GetUsers())
	setter.SetWorkflowIds(farmID, farm.GetWorkflows())
	return farm
}

func (setter *KeyValueSetter) SetDeviceIds(farmID uint64, devices []*config.Device) []*config.Device {
	for _, device := range devices {
		if device.GetID() == 0 {
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

func (setter *KeyValueSetter) SetDeviceSettingIds(device *config.Device, deviceSettings []*config.DeviceSetting) []*config.DeviceSetting {
	deviceID := device.GetID()
	for _, deviceSetting := range deviceSettings {
		if deviceSetting.GetID() == 0 {
			deviceSettingID := setter.idGenerator.NewDeviceSettingID(deviceID, deviceSetting.GetKey())
			deviceSetting.SetID(deviceSettingID)
		}
		setter.SetMetricIds(deviceID, device.GetMetrics())
		setter.SetChannelsIds(device.GetFarmID(), deviceID, device.GetChannels())
	}
	return deviceSettings
}

func (setter *KeyValueSetter) SetMetricIds(deviceID uint64, metrics []*config.Metric) []*config.Metric {
	for _, metric := range metrics {
		if metric.GetID() == 0 {
			metricID := setter.idGenerator.NewMetricID(deviceID, metric.GetKey())
			metric.SetID(metricID)
		}
		if metric.GetDeviceID() == 0 {
			metric.SetDeviceID(deviceID)
		}
	}
	return metrics
}

func (setter *KeyValueSetter) SetChannelsIds(farmID, deviceID uint64, channels []*config.Channel) []*config.Channel {
	for _, channel := range channels {
		if channel.GetID() == 0 {
			channelID := setter.idGenerator.NewChannelID(deviceID, channel.GetName())
			channel.SetID(channelID)
		}
		if channel.GetDeviceID() == 0 {
			channel.SetDeviceID(deviceID)
		}
		setter.SetConditionIds(channel.GetID(), deviceID, channel.GetConditions())
		setter.SetScheduleIds(farmID, deviceID, channel.GetID(), channel.GetSchedule())
	}
	return channels
}

func (setter *KeyValueSetter) SetConditionIds(channelID, deviceID uint64, conditions []*config.Condition) []*config.Condition {
	for _, condition := range conditions {
		if condition.GetID() == 0 {
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

func (setter *KeyValueSetter) SetScheduleIds(farmID, deviceID, channelID uint64, schedules []*config.Schedule) []*config.Schedule {
	for _, schedule := range schedules {
		if schedule.GetID() == 0 {
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

func (setter *KeyValueSetter) SetUserIds(users []*config.User) []*config.User {
	for _, user := range users {
		if user.GetID() == 0 {
			user.SetID(setter.idGenerator.NewUserID(user.GetEmail()))
		}
		setter.SetRoleIds(user.GetRoles())
	}
	return users
}

func (setter *KeyValueSetter) SetRoleIds(roles []*config.Role) []*config.Role {
	for _, role := range roles {
		if role.GetID() == 0 {
			role.SetID(setter.idGenerator.NewRoleID(role.GetName()))
		}
	}
	return roles
}

func (setter *KeyValueSetter) SetWorkflowIds(farmID uint64, workflows []*config.Workflow) []*config.Workflow {
	for _, workflow := range workflows {
		if workflow.GetID() == 0 {
			workflowID := setter.idGenerator.NewWorkflowID(farmID, workflow.GetName())
			workflow.SetID(workflowID)
		}
		setter.SetWorkflowStepIds(workflow.GetID(), workflow.GetSteps())
	}
	return workflows
}

func (setter *KeyValueSetter) SetWorkflowStepIds(workflowID uint64, workflowSteps []*config.WorkflowStep) []*config.WorkflowStep {
	for _, workflowStep := range workflowSteps {
		if workflowStep.GetID() == 0 {
			workflowStepKey := fmt.Sprintf("%d-%d-%d-%d-%d", workflowID, workflowStep.GetDeviceID(),
				workflowStep.GetChannelID(), workflowStep.GetDuration(), workflowStep.GetState())
			workflowStepID := setter.idGenerator.NewWorkflowStepID(workflowID, workflowStepKey)
			workflowStep.SetID(workflowStepID)
		}
		if workflowStep.GetWorkflowID() == 0 {
			workflowStep.SetID(workflowID)
		}
	}
	return workflowSteps
}
