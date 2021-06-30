package device

type DefaultDeviceInfo struct {
	HardwareVersion string `json:"hardware"`
	FirmwareVersion string `json:"firmware"`
	Uptime          int64  `json:"uptime"`
	DeviceInfo
}

func (di *DefaultDeviceInfo) GetHardwareVersion() string {
	return di.HardwareVersion
}

func (di *DefaultDeviceInfo) GetFirmwareVersion() string {
	return di.FirmwareVersion
}

func (di *DefaultDeviceInfo) GetUptime() int64 {
	return di.Uptime
}
