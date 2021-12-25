package provisioner

import (
	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
)

type ProvisionerParams struct {
	OrganizationID uint64
	ConfigStore    int
	StateStore     int
	DataStore      int
}

type FarmProvisioner interface {
	//BuildConfig(adminUser common.UserAccount) (config.FarmConfig, error)
	Provision(user common.UserAccount, params *ProvisionerParams) (config.FarmConfig, error)
	Deprovision(user common.UserAccount, farmID uint64) error
}

// const (
// 	DEFAULT_GALLONS = "50"

// 	CONFIG_ROOM_ENABLE_KEY = "room.enable"
// 	CONFIG_ROOM_NOTIFY_KEY = "room.notify"
// 	CONFIG_ROOM_URI_KEY    = "room.uri"
// 	CONFIG_ROOM_VIDEO_KEY  = "room.video"

// 	CONFIG_RESERVOIR_ENABLE_KEY             = "reservoir.enable"
// 	CONFIG_RESERVOIR_NOTIFY_KEY             = "reservoir.notify"
// 	CONFIG_RESERVOIR_URI_KEY                = "reservoir.uri"
// 	CONFIG_RESERVOIR_GALLONS_KEY            = "reservoir.gallons"
// 	CONFIG_RESERVOIR_TARGET_TEMP_KEY        = "reservoir.targetTemp"
// 	CONFIG_RESERVOIR_WATERCHANGE_ENABLE_KEY = "reservoir.waterchange.enable"
// 	CONFIG_RESERVOIR_WATERCHANGE_NOTIFY_KEY = "reservoir.waterchange.notify"

// 	CONFIG_DOSER_ENABLE_KEY    = "doser.enable"
// 	CONFIG_DOSER_NOTIFY_KEY    = "doser.notify"
// 	CONFIG_DOSER_URI_KEY       = "doser.uri"
// 	CONFIG_DOSER_GALLONS_KEY   = "doser.gallons"
// 	CONFIG_DOSER_CHANNEL_COUNT = 15

// 	METRIC_ROOM_MEMORY_KEY     = "mem"
// 	METRIC_ROOM_TEMPF0_KEY     = "tempF0"
// 	METRIC_ROOM_TEMPC0_KEY     = "tempC0"
// 	METRIC_ROOM_HUMIDITY0_KEY  = "humidity0"
// 	METRIC_ROOM_HEATINDEX0_KEY = "heatIndex0"
// 	METRIC_ROOM_TEMPF1_KEY     = "tempF1"
// 	METRIC_ROOM_TEMPC1_KEY     = "tempC1"
// 	METRIC_ROOM_HUMIDITY1_KEY  = "humidity1"
// 	METRIC_ROOM_HEATINDEX1_KEY = "heatIndex1"
// 	METRIC_ROOM_TEMPF2_KEY     = "tempF2"
// 	METRIC_ROOM_TEMPC2_KEY     = "tempC2"
// 	METRIC_ROOM_HUMIDITY2_KEY  = "humidity2"
// 	METRIC_ROOM_HEATINDEX2_KEY = "heatIndex2"
// 	METRIC_ROOM_WATERTEMP0_KEY = "water0"
// 	METRIC_ROOM_WATERTEMP1_KEY = "water1"
// 	METRIC_ROOM_VPD_KEY        = "vpd"
// 	METRIC_ROOM_CO2_KEY        = "co2"
// 	METRIC_ROOM_PHOTO_KEY      = "photo"
// 	METRIC_ROOM_WATERLEAK0_KEY = "leak0"
// 	METRIC_ROOM_WATERLEAK1_KEY = "leak1"

// 	METRIC_RESERVOIR_MEMORY_KEY       = "mem"
// 	METRIC_RESERVOIR_TEMP_KEY         = "resTemp"
// 	METRIC_RESERVOIR_PH_KEY           = "PH"
// 	METRIC_RESERVOIR_EC_KEY           = "EC"
// 	METRIC_RESERVOIR_TDS_KEY          = "TDS"
// 	METRIC_RESERVOIR_ORP_KEY          = "ORP"
// 	METRIC_RESERVOIR_DOMGL_KEY        = "DO_mgL"
// 	METRIC_RESERVOIR_DOPER_KEY        = "DO_PER"
// 	METRIC_RESERVOIR_SAL_KEY          = "SAL"
// 	METRIC_RESERVOIR_SG_KEY           = "SG"
// 	METRIC_RESERVOIR_ENVTEMP_KEY      = "envTemp"
// 	METRIC_RESERVOIR_ENVHUMIDITY_KEY  = "envHumidity"
// 	METRIC_RESERVOIR_ENVHEATINDEX_KEY = "envHeatIndex"
// 	METRIC_RESERVOIR_LOWERFLOAT_KEY   = "lowerFloat"
// 	METRIC_RESERVOIR_UPPERFLOAT_KEY   = "upperFloat"

// 	CHANNEL_ROOM_LIGHTING    = "Lighting"
// 	CHANNEL_ROOM_AC          = "Air Conditioner"
// 	CHANNEL_ROOM_HEATER      = "Heater"
// 	CHANNEL_ROOM_DEHUEY      = "Dehumidifier"
// 	CHANNEL_ROOM_VENTILATION = "Ventilation"
// 	CHANNEL_ROOM_CO2         = "Co2"

// 	CHANNEL_RESERVOIR_DRAIN     = "Drain"
// 	CHANNEL_RESERVOIR_CHILLER   = "Chiller"
// 	CHANNEL_RESERVOIR_HEATER    = "Heater"
// 	CHANNEL_RESERVOIR_POWERHEAD = "Powerhead"
// 	CHANNEL_RESERVOIR_AUX       = "Auxiliary"
// 	CHANNEL_RESERVOIR_TOPOFF    = "Top-off"
// 	CHANNEL_RESERVOIR_FAUCET    = "Faucet"

// 	CHANNEL_DOSER_PHDOWN   = "pH DOWN"
// 	CHANNEL_DOSER_PHUP     = "pH UP"
// 	CHANNEL_DOSER_OXIDIZER = "Oxidizer"
// 	CHANNEL_DOSER_TOPOFF   = "Top-off"
// 	CHANNEL_DOSER_NUTE1    = "Nutrient Part 1"
// 	CHANNEL_DOSER_NUTE2    = "Nutrient Part 2"
// 	CHANNEL_DOSER_NUTE3    = "Nutrient Part 3"
// )
