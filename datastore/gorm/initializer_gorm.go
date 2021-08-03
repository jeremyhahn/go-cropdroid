package gorm

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/entity"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
	"golang.org/x/crypto/bcrypt"
)

type GormInitializer struct {
	logger   *logging.Logger
	db       *gorm.DB
	location *time.Location
	farmDAO  dao.FarmDAO
	datastore.Initializer
}

func NewGormInitializer(logger *logging.Logger, gormDB GormDB,
	location *time.Location) datastore.Initializer {

	gormDB.Create()
	gormDB.Migrate()
	db := gormDB.GORM()
	return &GormInitializer{
		logger:   logger,
		db:       db,
		location: location,
		farmDAO:  NewFarmDAO(logger, db)}
}

// Initializes a new database, including a new administrative user and default FarmConfig.
func (initializer *GormInitializer) Initialize() error {

	encrypted, err := bcrypt.GenerateFromPassword([]byte(common.DEFAULT_PASSWORD), bcrypt.DefaultCost)
	if err != nil {
		initializer.logger.Fatalf("Error generating encrypted password: %s", err)
		return err
	}

	adminUser := config.NewUser()
	adminUser.SetEmail(common.DEFAULT_USER)
	adminUser.SetPassword(string(encrypted))

	initializer.db.Create(adminUser)

	farmConfig, err := initializer.BuildConfig(adminUser)
	if err != nil {
		return err
	}

	initializer.farmDAO.Save(farmConfig.(*config.Farm))

	initializer.seedInventory()

	return nil
}

// Builds a FarmConfig for the specified pre-existing admin user.
func (initializer *GormInitializer) BuildConfig(adminUser config.UserConfig) (config.FarmConfig, error) {

	defaultTimezone := initializer.location.String()
	now := time.Now().In(initializer.location)
	nowHr, nowMin, _ := now.Clock()
	sevenPM := time.Date(now.Year(), now.Month(), now.Day(), 19, 0, 0, 0, initializer.location)

	adminRole := config.NewRole()
	adminRole.SetName(common.DEFAULT_ROLE)
	initializer.db.Create(adminRole)
	initializer.db.Create(&config.Role{Name: "cultivator"})
	initializer.db.Create(&config.Role{Name: "analyst"})

	// common.SERVER_CONTROLLER_ID needs to match the ID of this server device!!
	serverDevice := config.NewDevice()
	serverDevice.SetType(common.CONTROLLER_TYPE_SERVER)
	serverDevice.SetDescription("Provides monitoring, real-time notifications, and web services")
	serverDevice.SetConfigs([]config.DeviceConfigItem{
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_NAME_KEY, Value: "First Room"},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_INTERVAL_KEY, Value: "60"},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_TIMEZONE_KEY, Value: defaultTimezone},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_MODE_KEY, Value: "virtual"},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_SMTP_ENABLE_KEY, Value: "false"},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_SMTP_HOST_KEY, Value: "smtp.gmail.com"},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_SMTP_PORT_KEY, Value: "587"},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_SMTP_USERNAME_KEY, Value: "myuser@gmail.com"},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_SMTP_PASSWORD_KEY, Value: "$ecret!"},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_SMTP_RECIPIENT_KEY, Value: "1234567890@vtext.com"}})

	roomDevice := config.NewDevice()
	roomDevice.SetType(common.CONTROLLER_TYPE_ROOM)
	roomDevice.SetDescription("Manages and monitors room climate")
	roomDevice.SetConfigs([]config.DeviceConfigItem{
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: roomDevice.GetID(), Key: common.CONFIG_ROOM_ENABLE_KEY, Value: "true"},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: roomDevice.GetID(), Key: common.CONFIG_ROOM_NOTIFY_KEY, Value: "true"},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: roomDevice.GetID(), Key: common.CONFIG_ROOM_URI_KEY},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: roomDevice.GetID(), Key: common.CONFIG_ROOM_VIDEO_KEY}})
	roomDevice.SetMetrics([]config.Metric{
		config.Metric{Key: common.METRIC_ROOM_MEMORY_KEY, Name: "Available System Memory", DataType: common.DATATYPE_INT, Unit: "bytes", Enable: true, Notify: true, AlarmLow: 500, AlarmHigh: 100000},
		config.Metric{Key: common.METRIC_ROOM_TEMPF0_KEY, Name: "Ceiling Air Temperature", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: true, AlarmLow: 71, AlarmHigh: 85},
		config.Metric{Key: common.METRIC_ROOM_HUMIDITY0_KEY, Name: "Ceiling Humidity", DataType: common.DATATYPE_FLOAT, Unit: "%", Enable: true, Notify: true, AlarmLow: 40, AlarmHigh: 70},
		config.Metric{Key: common.METRIC_ROOM_HEATINDEX0_KEY, Name: "Ceiling Heat Index", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: true, AlarmLow: 71, AlarmHigh: 85},
		config.Metric{Key: common.METRIC_ROOM_TEMPF1_KEY, Name: "Canopy Air Temperature", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: false, Notify: true, AlarmLow: 71, AlarmHigh: 85},
		config.Metric{Key: common.METRIC_ROOM_HUMIDITY1_KEY, Name: "Canopy Humidity", DataType: common.DATATYPE_FLOAT, Unit: "%", Enable: true, Notify: false, AlarmLow: 71, AlarmHigh: 85},
		config.Metric{Key: common.METRIC_ROOM_HEATINDEX1_KEY, Name: "Canopy Heat Index", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: false, AlarmLow: 71, AlarmHigh: 85},
		config.Metric{Key: common.METRIC_ROOM_TEMPF2_KEY, Name: "Floor Air Temperature", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: false, AlarmLow: 71, AlarmHigh: 85},
		config.Metric{Key: common.METRIC_ROOM_HUMIDITY2_KEY, Name: "Floor Humidity", DataType: common.DATATYPE_FLOAT, Unit: "%", Enable: true, Notify: false, AlarmLow: 71, AlarmHigh: 85},
		config.Metric{Key: common.METRIC_ROOM_HEATINDEX2_KEY, Name: "Floor Heat Index", DataType: common.DATATYPE_FLOAT, Unit: "bytes", Enable: true, Notify: false, AlarmLow: 71, AlarmHigh: 85},
		config.Metric{Key: common.METRIC_ROOM_WATERTEMP0_KEY, Name: "Pod 1 Water Temperature", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: true, AlarmLow: 61, AlarmHigh: 67},
		config.Metric{Key: common.METRIC_ROOM_WATERTEMP1_KEY, Name: "Pod 2 Water Temperature", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: true, AlarmLow: 61, AlarmHigh: 67},
		config.Metric{Key: common.METRIC_ROOM_VPD_KEY, Name: "Vapor Pressure Deficit", DataType: common.DATATYPE_FLOAT, Unit: "", Enable: true, Notify: false, AlarmLow: -2, AlarmHigh: 2},
		config.Metric{Key: common.METRIC_ROOM_CO2_KEY, Name: "Carbon Dioxide", DataType: common.DATATYPE_FLOAT, Unit: "ppm", Enable: true, Notify: false, AlarmLow: 800, AlarmHigh: 1300},
		config.Metric{Key: common.METRIC_ROOM_PHOTO_KEY, Name: "Light Sensor", DataType: common.DATATYPE_INT, Unit: "mV", Enable: true, Notify: false, AlarmLow: 800, AlarmHigh: 100},
		config.Metric{Key: common.METRIC_ROOM_WATERLEAK0_KEY, Name: "Pod 1 Water Leak", DataType: common.DATATYPE_INT, Unit: "mV", Enable: true, Notify: false, AlarmLow: -1, AlarmHigh: 1},
		config.Metric{Key: common.METRIC_ROOM_WATERLEAK1_KEY, Name: "Pod 2 Water Leak", DataType: common.DATATYPE_INT, Unit: "mV", Enable: true, Notify: false, AlarmLow: -1, AlarmHigh: 1}})
	ventOnHours := []int{13, 14, 15, 16, 17, 18}
	ventSchedules := make([]config.Schedule, len(ventOnHours))
	for i, hour := range ventOnHours {
		ventOn := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, initializer.location)
		ventSchedules[i] = config.Schedule{StartDate: ventOn, Frequency: common.SCHEDULE_FREQUENCY_DAILY}
	}
	roomDevice.SetChannels([]config.Channel{
		config.Channel{ChannelID: common.CHANNEL_ROOM_LIGHTING_ID, Name: common.CHANNEL_ROOM_LIGHTING, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 64800, AlgorithmID: 0,
			Schedule: []config.Schedule{config.Schedule{StartDate: sevenPM, Frequency: common.SCHEDULE_FREQUENCY_DAILY}}},
		config.Channel{ChannelID: common.CHANNEL_ROOM_AC_ID, Name: common.CHANNEL_ROOM_AC, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0},
		config.Channel{ChannelID: common.CHANNEL_ROOM_HEATER_ID, Name: common.CHANNEL_ROOM_HEATER, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0},
		config.Channel{ChannelID: common.CHANNEL_ROOM_DEHUEY_ID, Name: common.CHANNEL_ROOM_DEHUEY, Enable: true, Notify: true, Debounce: 10, Backoff: 0, Duration: 0, AlgorithmID: 0},
		config.Channel{ChannelID: common.CHANNEL_ROOM_VENTILATION_ID, Name: common.CHANNEL_ROOM_VENTILATION, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 900, AlgorithmID: 0,
			Schedule: ventSchedules},
		config.Channel{ChannelID: common.CHANNEL_ROOM_CO2_ID, Name: common.CHANNEL_ROOM_CO2, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0}})

	reservoirDevice := config.NewDevice()
	reservoirDevice.SetType(common.CONTROLLER_TYPE_RESERVOIR)
	reservoirDevice.SetDescription("Manages and monitors reservoir water and nutrients")
	reservoirDevice.SetConfigs([]config.DeviceConfigItem{
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: reservoirDevice.GetID(), Key: common.CONFIG_RESERVOIR_ENABLE_KEY, Value: "true"},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: reservoirDevice.GetID(), Key: common.CONFIG_RESERVOIR_NOTIFY_KEY, Value: "true"},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: reservoirDevice.GetID(), Key: common.CONFIG_RESERVOIR_URI_KEY},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: reservoirDevice.GetID(), Key: common.CONFIG_RESERVOIR_GALLONS_KEY, Value: common.DEFAULT_GALLONS},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: reservoirDevice.GetID(), Key: common.CONFIG_RESERVOIR_WATERCHANGE_ENABLE_KEY, Value: "false"},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: reservoirDevice.GetID(), Key: common.CONFIG_RESERVOIR_WATERCHANGE_NOTIFY_KEY, Value: "false"}})
	reservoirDevice.SetMetrics([]config.Metric{
		config.Metric{Key: common.METRIC_RESERVOIR_MEMORY_KEY, Name: "Available System Memory", DataType: common.DATATYPE_INT, Unit: "bytes", Enable: true, Notify: true, AlarmLow: 500, AlarmHigh: 100000},
		config.Metric{Key: common.METRIC_RESERVOIR_TEMP_KEY, Name: "Water Temperature", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: true, AlarmLow: 61, AlarmHigh: 67},
		config.Metric{Key: common.METRIC_RESERVOIR_PH_KEY, Name: "pH", DataType: common.DATATYPE_FLOAT, Unit: "", Enable: true, Notify: true, AlarmLow: 5.4, AlarmHigh: 6.2},
		config.Metric{Key: common.METRIC_RESERVOIR_EC_KEY, Name: "Electrical Conductivity (EC)", DataType: common.DATATYPE_FLOAT, Unit: "mV", Enable: true, Notify: true, AlarmLow: 850, AlarmHigh: 1300},
		config.Metric{Key: common.METRIC_RESERVOIR_TDS_KEY, Name: "Total Dissolved Solids (TDS)", DataType: common.DATATYPE_FLOAT, Unit: "ppm", Enable: true, Notify: false, AlarmLow: 700, AlarmHigh: 900},
		config.Metric{Key: common.METRIC_RESERVOIR_ORP_KEY, Name: "Oxygen Reduction Potential (ORP)", DataType: common.DATATYPE_FLOAT, Unit: "mV", Enable: true, Notify: false, AlarmLow: 250, AlarmHigh: 375},
		config.Metric{Key: common.METRIC_RESERVOIR_DOMGL_KEY, Name: "Dissolved Oxygen (DO)", DataType: common.DATATYPE_FLOAT, Unit: "mg/L", Enable: true, Notify: false, AlarmLow: 5, AlarmHigh: 30},
		config.Metric{Key: common.METRIC_RESERVOIR_DOPER_KEY, Name: "Dissolved Oxygen (DO)", DataType: common.DATATYPE_FLOAT, Unit: "%", Enable: true, Notify: false, AlarmLow: 0, AlarmHigh: 0},
		config.Metric{Key: common.METRIC_RESERVOIR_SAL_KEY, Name: "Salinity", DataType: common.DATATYPE_FLOAT, Unit: "ppt", Enable: true, Notify: false, AlarmLow: 0, AlarmHigh: 0},
		config.Metric{Key: common.METRIC_RESERVOIR_SG_KEY, Name: "Specific Gravity", DataType: common.DATATYPE_FLOAT, Unit: "ppt", Enable: true, Notify: false, AlarmLow: 0, AlarmHigh: 0},
		config.Metric{Key: common.METRIC_RESERVOIR_ENVTEMP_KEY, Name: "Environment Temp", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: false, AlarmLow: 40, AlarmHigh: 80},
		config.Metric{Key: common.METRIC_RESERVOIR_ENVHUMIDITY_KEY, Name: "Environment Humidity", DataType: common.DATATYPE_FLOAT, Unit: "%", Enable: true, Notify: false, AlarmLow: 40, AlarmHigh: 80},
		config.Metric{Key: common.METRIC_RESERVOIR_ENVHEATINDEX_KEY, Name: "Environment Heat Index", DataType: common.DATATYPE_FLOAT, Unit: "%", Enable: true, Notify: false, AlarmLow: 40, AlarmHigh: 80},
		config.Metric{Key: common.METRIC_RESERVOIR_LOWERFLOAT_KEY, Name: "Lower Float", DataType: common.DATATYPE_INT, Unit: "", Enable: true, Notify: false, AlarmLow: -1, AlarmHigh: 1},
		config.Metric{Key: common.METRIC_RESERVOIR_UPPERFLOAT_KEY, Name: "Upper Float", DataType: common.DATATYPE_INT, Unit: "", Enable: true, Notify: false, AlarmLow: -1, AlarmHigh: 1}})
	// Custom auxiliary schedule
	endDate := sevenPM.Add(time.Duration(common.HOURS_IN_A_YEAR) * time.Hour)
	days := "SU,TU,TH,SA"
	oneMinuteFromNow := time.Date(now.Year(), now.Month(), now.Day(), nowHr, nowMin+1, 0, 0, initializer.location)
	// Top-off schedule
	nineAM := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, initializer.location)
	ninePM := time.Date(now.Year(), now.Month(), now.Day(), 21, 0, 0, 0, initializer.location)
	reservoirDevice.SetChannels([]config.Channel{
		config.Channel{ChannelID: common.CHANNEL_RESERVOIR_DRAIN_ID, Name: common.CHANNEL_RESERVOIR_DRAIN, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0},
		config.Channel{ChannelID: common.CHANNEL_RESERVOIR_CHILLER_ID, Name: common.CHANNEL_RESERVOIR_CHILLER, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0},
		config.Channel{ChannelID: common.CHANNEL_RESERVOIR_HEATER_ID, Name: common.CHANNEL_RESERVOIR_HEATER, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0},
		config.Channel{ChannelID: common.CHANNEL_RESERVOIR_POWERHEAD_ID, Name: common.CHANNEL_RESERVOIR_POWERHEAD, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0},
		config.Channel{ChannelID: common.CHANNEL_RESERVOIR_AUX_ID, Name: common.CHANNEL_RESERVOIR_AUX, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 60, AlgorithmID: 0,
			Schedule: []config.Schedule{
				config.Schedule{StartDate: sevenPM, EndDate: &endDate, Frequency: common.SCHEDULE_FREQUENCY_WEEKLY, Interval: 2, Count: 5, Days: &days},
				config.Schedule{StartDate: oneMinuteFromNow, Frequency: common.SCHEDULE_FREQUENCY_DAILY}}},
		config.Channel{ChannelID: common.CHANNEL_RESERVOIR_TOPOFF_ID, Name: common.CHANNEL_RESERVOIR_TOPOFF, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 120, AlgorithmID: 0,
			Schedule: []config.Schedule{
				config.Schedule{StartDate: nineAM, Frequency: common.SCHEDULE_FREQUENCY_WEEKLY},
				config.Schedule{StartDate: ninePM, Frequency: common.SCHEDULE_FREQUENCY_MONTHLY},
				config.Schedule{StartDate: ninePM, Frequency: common.SCHEDULE_FREQUENCY_YEARLY}}},
		config.Channel{ChannelID: common.CHANNEL_RESERVOIR_FAUCET_ID, Name: common.CHANNEL_RESERVOIR_FAUCET, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0}})

	doserDevice := config.NewDevice()
	doserDevice.SetType(common.CONTROLLER_TYPE_DOSER)
	doserDevice.SetDescription("Nutrient dosing and expansion I/O device")
	doserDevice.SetConfigs([]config.DeviceConfigItem{
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: doserDevice.GetID(), Key: common.CONFIG_DOSER_ENABLE_KEY, Value: "true"},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: doserDevice.GetID(), Key: common.CONFIG_DOSER_NOTIFY_KEY, Value: "true"},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: doserDevice.GetID(), Key: common.CONFIG_DOSER_URI_KEY},
		config.DeviceConfigItem{UserID: adminUser.GetID(), DeviceID: doserDevice.GetID(), Key: common.CONFIG_DOSER_GALLONS_KEY, Value: common.DEFAULT_GALLONS}})
	doserDevice.SetChannels([]config.Channel{
		config.Channel{ChannelID: common.CHANNEL_DOSER_PHDOWN_ID, Name: common.CHANNEL_DOSER_PHDOWN, Enable: true, Notify: true, Debounce: 0, Backoff: 10, Duration: 0, AlgorithmID: 1},
		config.Channel{ChannelID: common.CHANNEL_DOSER_PHUP_ID, Name: common.CHANNEL_DOSER_PHUP, Enable: false, Notify: true, Debounce: 0, Backoff: 10, Duration: 0, AlgorithmID: 1},
		config.Channel{ChannelID: common.CHANNEL_DOSER_OXIDIZER_ID, Name: common.CHANNEL_DOSER_OXIDIZER, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 2},
		config.Channel{ChannelID: common.CHANNEL_DOSER_TOPOFF_ID, Name: common.CHANNEL_DOSER_TOPOFF, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 120, AlgorithmID: 0},
		config.Channel{ChannelID: common.CHANNEL_DOSER_NUTE1_ID, Name: common.CHANNEL_DOSER_NUTE1, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 30, AlgorithmID: 0},
		config.Channel{ChannelID: common.CHANNEL_DOSER_NUTE2_ID, Name: common.CHANNEL_DOSER_NUTE2, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 30, AlgorithmID: 0},
		config.Channel{ChannelID: common.CHANNEL_DOSER_NUTE3_ID, Name: common.CHANNEL_DOSER_NUTE3, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 30, AlgorithmID: 0}})

	farm := config.NewFarm()
	farm.Consistency = common.CONSISTENCY_CACHED
	farm.SetDevices([]config.Device{*serverDevice, *roomDevice, *reservoirDevice, *doserDevice})
	initializer.farmDAO.Create(farm)

	initializer.db.Create(&config.Permission{
		UserID: adminUser.GetID(),
		RoleID: adminRole.GetID(),
		FarmID: farm.GetID()})

	initializer.db.Create(&config.Algorithm{Name: "pH"})
	initializer.db.Create(&config.Algorithm{Name: "Oxidizer"})

	// Create conditions now that metric ids have been saved to the databsae.
	// GORM has trouble managing tables  with multiple foreign keys
	farmConfig, err := initializer.farmDAO.Get(farm.GetID())
	if err != nil {
		return nil, err
	}

	devices := farmConfig.GetDevices()
	persistedRoom := devices[1]
	persistedReservoir := devices[2]
	persistedDoser := devices[3]
	roomChannels := persistedRoom.GetChannels()
	reservoirChannels := persistedReservoir.GetChannels()
	doserChannels := persistedDoser.GetChannels()

	roomChannels[1].SetConditions([]config.Condition{
		config.Condition{
			MetricID:   roomDevice.GetMetrics()[1].GetID(),
			Comparator: ">",
			Threshold:  85.0}})

	roomChannels[2].SetConditions([]config.Condition{
		config.Condition{
			MetricID:   roomDevice.GetMetrics()[1].GetID(),
			Comparator: "<",
			Threshold:  70.0}})

	roomChannels[3].SetConditions([]config.Condition{
		config.Condition{
			MetricID:   roomDevice.GetMetrics()[3].GetID(),
			Comparator: ">",
			Threshold:  55.0}})

	roomChannels[5].SetConditions([]config.Condition{
		config.Condition{
			MetricID:   roomDevice.GetMetrics()[14].GetID(),
			Comparator: "<",
			Threshold:  1200.0}})

	reservoirChannels[1].SetConditions([]config.Condition{
		config.Condition{
			MetricID:   reservoirDevice.GetMetrics()[2].GetID(),
			Comparator: ">",
			Threshold:  62.0}})

	reservoirChannels[2].SetConditions([]config.Condition{
		config.Condition{
			MetricID:   reservoirDevice.GetMetrics()[2].GetID(),
			Comparator: "<",
			Threshold:  60.0}})

	doserChannels[0].SetConditions([]config.Condition{
		config.Condition{
			MetricID:   reservoirDevice.GetMetrics()[2].GetID(),
			Comparator: ">",
			Threshold:  6.1}})

	doserChannels[1].SetConditions([]config.Condition{
		config.Condition{
			MetricID:   reservoirDevice.GetMetrics()[2].GetID(),
			Comparator: "<",
			Threshold:  5.4}})

	doserChannels[2].SetConditions([]config.Condition{
		config.Condition{
			MetricID:   reservoirDevice.GetMetrics()[5].GetID(),
			Comparator: "<",
			Threshold:  300.0}})

	// Water change workflow
	drainStep := config.NewWorkflowStep()
	drainStep.SetDeviceID(persistedReservoir.GetID())
	drainStep.SetChannelID(7)  // CHANNEL_RESERVOIR_DRAIN
	drainStep.SetDuration(300) // seconds; 5 minutes
	drainStep.SetWait(300)     // seconds; 5 minutes
	// drainStep.SetDuration(5)
	// drainStep.SetWait(10)

	fillStep := config.NewWorkflowStep()
	fillStep.SetDeviceID(persistedReservoir.GetID())
	fillStep.SetChannelID(13) // CHANNEL_RESERVOIR_FAUCET
	fillStep.SetDuration(300) // seconds; 5 minutes
	fillStep.SetWait(300)     // seconds; 5 minutes
	// fillStep.SetDuration(5)
	// fillStep.SetWait(10)

	phDownStep := config.NewWorkflowStep()
	phDownStep.SetDeviceID(persistedDoser.GetID())
	phDownStep.SetChannelID(14) // CHANNEL_DOSER_PHDOWN
	phDownStep.SetDuration(60)  // seconds
	phDownStep.SetWait(300)     // seconds; 5 minutes
	// phDownStep.SetDuration(5)
	// phDownStep.SetWait(10)

	nutePart1Step := config.NewWorkflowStep()
	nutePart1Step.SetDeviceID(persistedDoser.GetID())
	nutePart1Step.SetChannelID(18) // CHANNEL_DOSER_NUTE1
	nutePart1Step.SetDuration(30)  // seconds
	nutePart1Step.SetWait(300)     // seconds; 5 minutes
	// nutePart1Step.SetDuration(5)
	// nutePart1Step.SetWait(10)

	nutePart2Step := config.NewWorkflowStep()
	nutePart2Step.SetDeviceID(persistedDoser.GetID())
	nutePart2Step.SetChannelID(19) // CHANNEL_DOSER_NUTE2
	nutePart2Step.SetDuration(30)  // seconds
	nutePart2Step.SetWait(300)     // seconds; 5 minutes
	// nutePart2Step.SetDuration(5)
	// nutePart2Step.SetWait(10)

	nutePart3Step := config.NewWorkflowStep()
	nutePart3Step.SetDeviceID(persistedDoser.GetID())
	nutePart3Step.SetChannelID(20) // CHANNEL_DOSER_NUTE3
	nutePart3Step.SetDuration(30)  // seconds
	nutePart3Step.SetWait(300)     // seconds; 5 minutes
	// nutePart3Step.SetDuration(5)
	// nutePart3Step.SetWait(10)

	workflow1 := config.NewWorkflow()
	workflow1.SetName("Automated Water Changes")
	workflow1.SetSteps([]config.WorkflowStep{
		*drainStep,
		*fillStep,
		*phDownStep,
		*nutePart1Step,
		*nutePart2Step,
		*nutePart3Step})

	farmConfig.SetWorkflows([]config.Workflow{*workflow1})

	return farmConfig, nil
}

// Initializes the initial product inventory
func (initializer *GormInitializer) seedInventory() {
	initializer.db.Create(&entity.InventoryType{
		Name:             "pH Probe",
		Description:      "+/- 0.1 resolution",
		Image:            "https://www.atlas-scientific.com/_images/large_images/probes/c-ph-box.png",
		LifeExpectancy:   47336400, // 1.5 yrs
		MaintenanceCycle: 31557600, // 1 yr
		ProductPage:      "https://www.atlas-scientific.com/product_pages/probes/c-ph-probe.html"})
	initializer.db.Create(&entity.InventoryType{
		Name:             "Oxygen Reduction Potential Probe (ORP)",
		Description:      "+/– 1.1mV accuracy",
		Image:            "https://www.atlas-scientific.com/_images/large_images/probes/c-orp-box.png",
		MaintenanceCycle: 31557600, // 1 yr
		LifeExpectancy:   47336400, // 1.5 yrs
		ProductPage:      "https://www.atlas-scientific.com/product_pages/probes/c-orp-probe.html"})
	initializer.db.Create(&entity.InventoryType{
		Name:             "Conductivity Probe (EC/TDS)",
		Description:      "+/ – 2% accuracy",
		Image:            "https://www.atlas-scientific.com/_images/large_images/probes/ec_k0-1.png",
		LifeExpectancy:   315576000, // 10 yrs
		MaintenanceCycle: 315576000,
		ProductPage:      "https://www.atlas-scientific.com/product_pages/probes/ec_k0-1.html"})
	initializer.db.Create(&entity.InventoryType{
		Name:             "Dissolved Oxygen Probe (DO)",
		Description:      "+/ – 2% accuracy",
		Image:            "https://www.atlas-scientific.com/_images/large_images/probes/do_box.png",
		LifeExpectancy:   157788000, // 5 yrs
		MaintenanceCycle: 47336400,  // 18 months
		ProductPage:      "https://www.atlas-scientific.com/product_pages/probes/do_probe.html"})
	initializer.db.Create(&entity.InventoryType{
		Name:           "NDIR Carbon Dioxide Sensor (CO2)",
		Description:    "+/- 3%  +/- 30 ppm",
		Image:          "https://dqzrr9k4bjpzk.cloudfront.net/images/805759/897784873.jpg",
		LifeExpectancy: 173566800, // 5.5 yrs
		ProductPage:    "https://www.atlas-scientific.com/product_pages/probes/ezo-co2.html"})
	initializer.db.Create(&entity.InventoryType{
		Name:           "pH Calibration Fluids",
		Description:    "3x NIST reference calibrated 500ml bottles (4.0, 7.0, 10.0)",
		Image:          "https://images-na.ssl-images-amazon.com/images/I/81HsYv4kNTL._SL1500_.jpg",
		LifeExpectancy: 31557600, // 1 yr
		ProductPage:    "https://www.amazon.com/Biopharm-Calibration-Solution-Standards-Traceable/dp/B01E7U873K"})
	initializer.db.Create(&entity.InventoryType{
		Name:           "pH UP (BASE)",
		Description:    "General Hydroponics pH Up Liquid Fertilizer, 1-Gallon",
		Image:          "https://images-na.ssl-images-amazon.com/images/I/61mHEr-obpL._AC_SL1200_.jpg",
		LifeExpectancy: 31557600, // 1 yr
		ProductPage:    "https://www.amazon.com/General-Hydroponics-Liquid-Fertilizer-1-Gallon/dp/B000FF5V90"})
	initializer.db.Create(&entity.InventoryType{
		Name:           "pH DOWN (ACID)",
		Description:    "General Hydroponics HGC722125 Liquid Premium Buffering for pH Stability, 1-Gallon, Orange",
		Image:          "https://images-na.ssl-images-amazon.com/images/I/71E-fJ-tlsL._AC_SL1500_.jpg",
		LifeExpectancy: 31557600, // 1 yr
		ProductPage:    "https://www.amazon.com/General-Hydroponics-Liquid-Fertilizer-1-Gallon/dp/B000FG0F9U"})
}
