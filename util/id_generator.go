package util

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/entity"
)

type IdGenerator interface {
	NewID(string) uint64
	NewID32(str string) int
	Uint64Bytes(uint64) []byte
	StringBytes(str string) []byte
	TimestampBytes(t time.Time) []byte

	NewFarmID(orgID uint64, farmName string) uint64
	NewDeviceID(farmID uint64, deviceType string) uint64
	NewDeviceDataID(farmID, deviceID uint64, timestamp time.Time) uint64
	NewDeviceSettingID(deviceID uint64, deviceSettingKey string) uint64
	NewMetricID(deviceID uint64, metricKey string) uint64
	NewChannelID(deviceID uint64, channelName string) uint64
	NewConditionID(deviceID uint64, conditionKey string) uint64
	NewScheduleID(deviceID uint64, scheduleKey string) uint64
	NewUserID(email string) uint64
	NewRoleID(name string) uint64
	NewWorkflowID(farmID uint64, workflowName string) uint64
	NewWorkflowStepID(workflowID uint64, workflowStepKey string) uint64
	NewEventLogID(eventLog entity.EventLog) uint64

	CreateEventLogClusterID(clusterID uint64) uint64
	CreateDeviceDataClusterID(deviceID uint64) uint64
}

type Fnv1aHasher struct {
	is64bit bool
	IdGenerator
}

func NewIdGenerator(datastoreEngine string) IdGenerator {
	uid := &Fnv1aHasher{}
	if datastoreEngine == common.DATASTORE_TYPE_64BIT {
		// TODO: Fix this to true
		uid.is64bit = false
	} else {
		uid.is64bit = false
	}
	return uid
}

// Returns a new 64-bit FNV-1a hash from a string
func (hasher *Fnv1aHasher) NewID(str string) uint64 {
	return hasher.createClusterHash([]byte(str))
}

// Returns a new 64-bit FNV-1a hash from a string
func (hasher *Fnv1aHasher) NewID32(str string) int {
	fnv32a := fnv.New32a()
	fnv32a.Write([]byte(str))
	return int(fnv32a.Sum32())
}

// Returns a new 64-bit FNV-1a hash from a byte array
func (hasher *Fnv1aHasher) createClusterHash(bytes []byte) uint64 {
	if hasher.is64bit {
		clusterHash := fnv.New64a()
		clusterHash.Write(bytes)
		return clusterHash.Sum64()
	}
	clusterHash := fnv.New32a()
	clusterHash.Write(bytes)
	return uint64(clusterHash.Sum32())
}

// Returns a new 64-bit FNV-1a hash as little endian byte array
func (hasher *Fnv1aHasher) Uint64Bytes(id uint64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, id)
	return bytes
}

func (hasher *Fnv1aHasher) StringBytes(str string) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, hasher.NewID(str))
	return bytes
}

// Returns a new 64-bit FNV-1a hash from a time object and farm id
func (hasher *Fnv1aHasher) TimestampBytes(t time.Time) []byte {
	return hasher.Uint64Bytes(uint64(t.Unix()))
}

// Implementation specific ID hashes
func (hasher *Fnv1aHasher) NewFarmID(orgID uint64, farmName string) uint64 {
	return hasher.NewID(fmt.Sprintf("%d-%s", orgID, farmName))
}

func (hasher *Fnv1aHasher) NewDeviceID(farmID uint64, deviceType string) uint64 {
	return hasher.NewID(fmt.Sprintf("%d-%s", farmID, deviceType))
}

func (hasher *Fnv1aHasher) NewDeviceDataID(farmID, deviceID uint64, timestamp time.Time) uint64 {
	return hasher.NewID(fmt.Sprintf("%d-%d-%d", farmID, deviceID, timestamp.Unix()))
}

func (hasher *Fnv1aHasher) NewDeviceSettingID(deviceID uint64, deviceSettingKey string) uint64 {
	return hasher.NewID(fmt.Sprintf("%d-%s", deviceID, deviceSettingKey))
}

func (hasher *Fnv1aHasher) NewMetricID(deviceID uint64, metricKey string) uint64 {
	return hasher.NewID(fmt.Sprintf("%d-%s", deviceID, metricKey))
}

func (hasher *Fnv1aHasher) NewChannelID(deviceID uint64, channelName string) uint64 {
	return hasher.NewID(fmt.Sprintf("%d-%s", deviceID, channelName))
}

func (hasher *Fnv1aHasher) NewConditionID(deviceID uint64, conditionKey string) uint64 {
	return hasher.NewID(fmt.Sprintf("%d-%s", deviceID, conditionKey))
}

func (hasher *Fnv1aHasher) NewScheduleID(deviceID uint64, scheduleKey string) uint64 {
	return hasher.NewID(fmt.Sprintf("%d-%s", deviceID, scheduleKey))
}

func (hasher *Fnv1aHasher) NewEventLogID(eventLog entity.EventLog) uint64 {
	return hasher.NewID(fmt.Sprintf("%d-%d-%s-%d",
		eventLog.FarmID,
		eventLog.DeviceID,
		eventLog.Message,
		eventLog.Timestamp.Unix()))
}

func (hasher *Fnv1aHasher) NewUserID(email string) uint64 {
	return hasher.NewID(email)
}

func (hasher *Fnv1aHasher) NewRoleID(name string) uint64 {
	return hasher.NewID(name)
}

func (hasher *Fnv1aHasher) NewWorkflowID(farmID uint64, workflowName string) uint64 {
	return hasher.NewID(fmt.Sprintf("%d-%s", farmID, workflowName))
}

func (hasher *Fnv1aHasher) NewWorkflowStepID(workflowID uint64, workflowStepKey string) uint64 {
	return hasher.NewID(fmt.Sprintf("%d-%s", workflowID, workflowStepKey))
}

// Implementation specific cluster ID generation functions

func (hasher *Fnv1aHasher) CreateEventLogClusterID(clusterID uint64) uint64 {
	eventLogClusterID := hasher.NewID(fmt.Sprintf("%d-%s", clusterID, "eventlog"))
	fmt.Println(fmt.Sprintf("Creating event log cluster ID for clusterID: %d, eventLogClusterID=%d",
		clusterID, eventLogClusterID))
	return eventLogClusterID
}

func (hasher *Fnv1aHasher) CreateDeviceDataClusterID(deviceID uint64) uint64 {
	deviceDataClusterID := hasher.NewID(fmt.Sprintf("%d-%s", deviceID, "devicedata"))
	fmt.Println(fmt.Sprintf("Creating device data cluster ID for deviceID:%d, deviceDataClusterID=%d",
		deviceID, deviceDataClusterID))
	return deviceDataClusterID
}
