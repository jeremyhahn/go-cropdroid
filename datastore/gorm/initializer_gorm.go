package gorm

import (
	"fmt"
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/entity"
	"github.com/jeremyhahn/go-cropdroid/state"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
	"golang.org/x/crypto/bcrypt"
)

type GormInitializer struct {
	logger        *logging.Logger
	gormDB        GormDB
	db            *gorm.DB
	location      *time.Location
	farmDAO       dao.FarmDAO
	userDAO       dao.UserDAO
	roleDAO       dao.RoleDAO
	permissionDAO dao.PermissionDAO
	idGenerator   util.IdGenerator
	appMode       string
	datastore.Initializer
}

func NewGormInitializer(logger *logging.Logger, gormDB GormDB,
	idGenerator util.IdGenerator, location *time.Location,
	appMode string) datastore.Initializer {

	db := gormDB.GORM()
	return &GormInitializer{
		logger:        logger,
		db:            db,
		gormDB:        gormDB,
		idGenerator:   idGenerator,
		location:      location,
		farmDAO:       NewFarmDAO(logger, db),
		userDAO:       NewUserDAO(logger, db),
		roleDAO:       NewRoleDAO(logger, db),
		permissionDAO: NewPermissionDAO(logger, db),
		appMode:       appMode}
}

// Initializes a new database, including a new administrative user and default FarmConfig.
func (initializer *GormInitializer) Initialize(includeFarmConfig bool) error {

	initializer.gormDB.Create()
	initializer.gormDB.Migrate()

	encrypted, err := bcrypt.GenerateFromPassword([]byte(common.DEFAULT_PASSWORD), bcrypt.DefaultCost)
	if err != nil {
		initializer.logger.Fatalf("Error generating encrypted password: %s", err)
		return err
	}

	adminRole := config.NewRole()
	adminRole.SetID(initializer.newID(common.DEFAULT_ROLE))
	adminRole.SetName(common.DEFAULT_ROLE)
	initializer.roleDAO.Create(adminRole)

	cultivatorRole := config.NewRole()
	cultivatorRole.SetID(initializer.newID(common.ROLE_CULTIVATOR))
	cultivatorRole.SetName(common.ROLE_CULTIVATOR)
	initializer.roleDAO.Create(cultivatorRole)

	analystRole := config.NewRole()
	analystRole.SetID(initializer.newID(common.ROLE_ANALYST))
	analystRole.SetName(common.ROLE_ANALYST)
	initializer.roleDAO.Create(analystRole)

	adminUser := config.NewUser()
	adminUser.ID = initializer.newID(common.DEFAULT_USER)
	adminUser.SetEmail(common.DEFAULT_USER)
	adminUser.SetPassword(string(encrypted))
	initializer.userDAO.Create(adminUser)

	permission := &config.Permission{
		UserID: adminUser.GetID(),
		RoleID: adminRole.GetID()}
	initializer.permissionDAO.Save(permission)

	if includeFarmConfig {
		farmConfig, err := initializer.BuildConfig(adminUser, adminRole)
		if err != nil {
			return err
		}
		initializer.farmDAO.Save(farmConfig.(*config.Farm))
	}

	phAlgoID := initializer.newID(common.ALGORITHM_PH_KEY)
	phAlgo := &config.Algorithm{ID: phAlgoID, Name: common.ALGORITHM_PH_KEY}
	initializer.db.Create(phAlgo)

	oxidizerAlgoID := initializer.newID(common.ALGORITHM_ORP_KEY)
	oxidizerAlgo := &config.Algorithm{ID: oxidizerAlgoID, Name: common.ALGORITHM_ORP_KEY}
	initializer.db.Create(oxidizerAlgo)

	initializer.seedInventory()

	return nil
}

func (initializer *GormInitializer) newID(key string) uint64 {
	return initializer.idGenerator.NewID(key)
}

func (initializer *GormInitializer) newFarmID(farmID uint64, key string) uint64 {
	return initializer.idGenerator.NewID(fmt.Sprintf("%d-%s", farmID, key))
}

// Builds a FarmConfig for the specified pre-existing admin user.
func (initializer *GormInitializer) BuildConfig(adminUser config.UserConfig, assignedRole config.RoleConfig) (config.FarmConfig, error) {

	farmName := "My Crop"
	farmKey := fmt.Sprintf("%d-%s", adminUser.GetID(), farmName)
	farmID := initializer.idGenerator.NewID(farmKey)

	defaultTimezone := initializer.location.String()
	now := time.Now().In(initializer.location)
	//nowHr, nowMin, _ := now.Clock()
	sevenPM := time.Date(now.Year(), now.Month(), now.Day(), 19, 0, 0, 0, initializer.location)

	// common.SERVER_CONTROLLER_ID needs to match the ID of this server device!!

	serverDevice := config.NewDevice()
	serverDeviceNameID := initializer.newFarmID(farmID, "server-name")
	serverDeviceIntervalID := initializer.newFarmID(farmID, "server-interval")
	serverDeviceTimezoneID := initializer.newFarmID(farmID, "server-timezone")
	serverDeviceModeID := initializer.newFarmID(farmID, "server-mode")
	serverDeviceSmtpEnableID := initializer.newFarmID(farmID, "server-smtp-enable")
	serverDeviceSmtpHostID := initializer.newFarmID(farmID, "server-smtp-host")
	serverDeviceSmtpPortID := initializer.newFarmID(farmID, "server-smtp-port")
	serverDeviceSmtpUsernameID := initializer.newFarmID(farmID, "server-smtp-username")
	serverDeviceSmtpPasswordID := initializer.newFarmID(farmID, "server-smtp-password")
	serverDeviceSmtpRecipientID := initializer.newFarmID(farmID, "server-smtp-recipient")
	serverDeviceID := initializer.newFarmID(farmID, common.CONTROLLER_TYPE_SERVER)
	serverDevice.SetID(serverDeviceID)
	serverDevice.SetType(common.CONTROLLER_TYPE_SERVER)
	serverDevice.SetDescription("Provides monitoring, real-time notifications, and web services")
	serverDevice.SetConfigs([]config.DeviceConfigItem{
		{ID: serverDeviceNameID, UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_NAME_KEY, Value: "First Room"},
		{ID: serverDeviceIntervalID, UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_INTERVAL_KEY, Value: "60"},
		{ID: serverDeviceTimezoneID, UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_TIMEZONE_KEY, Value: defaultTimezone},
		{ID: serverDeviceModeID, UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_MODE_KEY, Value: "virtual"},
		{ID: serverDeviceSmtpEnableID, UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_SMTP_ENABLE_KEY, Value: "false"},
		{ID: serverDeviceSmtpHostID, UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_SMTP_HOST_KEY, Value: "smtp.gmail.com"},
		{ID: serverDeviceSmtpPortID, UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_SMTP_PORT_KEY, Value: "587"},
		{ID: serverDeviceSmtpUsernameID, UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_SMTP_USERNAME_KEY, Value: "myuser@gmail.com"},
		{ID: serverDeviceSmtpPasswordID, UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_SMTP_PASSWORD_KEY, Value: "$ecret!"},
		{ID: serverDeviceSmtpRecipientID, UserID: adminUser.GetID(), DeviceID: serverDevice.GetID(), Key: common.CONFIG_SMTP_RECIPIENT_KEY, Value: "1234567890@vtext.com"}})

	roomDevice := config.NewDevice()
	roomDeviceEnableID := initializer.newFarmID(farmID, "room-enable")
	roomDeviceNotifyID := initializer.newFarmID(farmID, "room-notify")
	roomDeviceUriID := initializer.newFarmID(farmID, "room-uri")
	roomDeviceVideoID := initializer.newFarmID(farmID, "room-video")
	roomDeviceMetric0ID := initializer.newFarmID(farmID, "metric-0")
	roomDeviceMetric1ID := initializer.newFarmID(farmID, "metric-1")
	roomDeviceMetric2ID := initializer.newFarmID(farmID, "metric-2")
	roomDeviceMetric3ID := initializer.newFarmID(farmID, "metric-3")
	roomDeviceMetric4ID := initializer.newFarmID(farmID, "metric-4")
	roomDeviceMetric5ID := initializer.newFarmID(farmID, "metric-5")
	roomDeviceMetric6ID := initializer.newFarmID(farmID, "metric-6")
	roomDeviceMetric7ID := initializer.newFarmID(farmID, "metric-7")
	roomDeviceMetric8ID := initializer.newFarmID(farmID, "metric-8")
	roomDeviceMetric9ID := initializer.newFarmID(farmID, "metric-9")
	roomDeviceMetric10ID := initializer.newFarmID(farmID, "metric-10")
	roomDeviceMetric11ID := initializer.newFarmID(farmID, "metric-11")
	roomDeviceMetric12ID := initializer.newFarmID(farmID, "metric-12")
	roomDeviceMetric13ID := initializer.newFarmID(farmID, "metric-13")
	roomDeviceMetric14ID := initializer.newFarmID(farmID, "metric-14")
	roomDeviceMetric15ID := initializer.newFarmID(farmID, "metric-15")
	roomDeviceMetric16ID := initializer.newFarmID(farmID, "metric-16")
	roomDevice.SetID(initializer.newFarmID(farmID, common.CONTROLLER_TYPE_ROOM))
	roomDevice.SetType(common.CONTROLLER_TYPE_ROOM)
	roomDevice.SetDescription("Manages and monitors room climate")
	roomDevice.SetConfigs([]config.DeviceConfigItem{
		{ID: roomDeviceEnableID, UserID: adminUser.GetID(), DeviceID: roomDevice.GetID(), Key: common.CONFIG_ROOM_ENABLE_KEY, Value: "true"},
		{ID: roomDeviceNotifyID, UserID: adminUser.GetID(), DeviceID: roomDevice.GetID(), Key: common.CONFIG_ROOM_NOTIFY_KEY, Value: "true"},
		{ID: roomDeviceUriID, UserID: adminUser.GetID(), DeviceID: roomDevice.GetID(), Key: common.CONFIG_ROOM_URI_KEY},
		{ID: roomDeviceVideoID, UserID: adminUser.GetID(), DeviceID: roomDevice.GetID(), Key: common.CONFIG_ROOM_VIDEO_KEY}})
	roomDevice.SetMetrics([]config.Metric{
		{ID: roomDeviceMetric0ID, Key: common.METRIC_ROOM_MEMORY_KEY, Name: "Available System Memory", DataType: common.DATATYPE_INT, Unit: "bytes", Enable: true, Notify: true, AlarmLow: 500, AlarmHigh: 100000},
		{ID: roomDeviceMetric1ID, Key: common.METRIC_ROOM_TEMPF0_KEY, Name: "Ceiling Air Temperature", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: true, AlarmLow: 71, AlarmHigh: 85},
		{ID: roomDeviceMetric2ID, Key: common.METRIC_ROOM_HUMIDITY0_KEY, Name: "Ceiling Humidity", DataType: common.DATATYPE_FLOAT, Unit: "%", Enable: true, Notify: true, AlarmLow: 40, AlarmHigh: 70},
		{ID: roomDeviceMetric3ID, Key: common.METRIC_ROOM_HEATINDEX0_KEY, Name: "Ceiling Heat Index", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: true, AlarmLow: 71, AlarmHigh: 85},
		{ID: roomDeviceMetric4ID, Key: common.METRIC_ROOM_TEMPF1_KEY, Name: "Canopy Air Temperature", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: false, Notify: true, AlarmLow: 71, AlarmHigh: 85},
		{ID: roomDeviceMetric5ID, Key: common.METRIC_ROOM_HUMIDITY1_KEY, Name: "Canopy Humidity", DataType: common.DATATYPE_FLOAT, Unit: "%", Enable: true, Notify: false, AlarmLow: 71, AlarmHigh: 85},
		{ID: roomDeviceMetric6ID, Key: common.METRIC_ROOM_HEATINDEX1_KEY, Name: "Canopy Heat Index", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: false, AlarmLow: 71, AlarmHigh: 85},
		{ID: roomDeviceMetric7ID, Key: common.METRIC_ROOM_TEMPF2_KEY, Name: "Floor Air Temperature", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: false, AlarmLow: 71, AlarmHigh: 85},
		{ID: roomDeviceMetric8ID, Key: common.METRIC_ROOM_HUMIDITY2_KEY, Name: "Floor Humidity", DataType: common.DATATYPE_FLOAT, Unit: "%", Enable: true, Notify: false, AlarmLow: 71, AlarmHigh: 85},
		{ID: roomDeviceMetric9ID, Key: common.METRIC_ROOM_HEATINDEX2_KEY, Name: "Floor Heat Index", DataType: common.DATATYPE_FLOAT, Unit: "bytes", Enable: true, Notify: false, AlarmLow: 71, AlarmHigh: 85},
		{ID: roomDeviceMetric10ID, Key: common.METRIC_ROOM_WATERTEMP0_KEY, Name: "Pod 1 Water Temperature", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: true, AlarmLow: 61, AlarmHigh: 67},
		{ID: roomDeviceMetric11ID, Key: common.METRIC_ROOM_WATERTEMP1_KEY, Name: "Pod 2 Water Temperature", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: true, AlarmLow: 61, AlarmHigh: 67},
		{ID: roomDeviceMetric12ID, Key: common.METRIC_ROOM_VPD_KEY, Name: "Vapor Pressure Deficit", DataType: common.DATATYPE_FLOAT, Unit: "", Enable: true, Notify: false, AlarmLow: -2, AlarmHigh: 2},
		{ID: roomDeviceMetric13ID, Key: common.METRIC_ROOM_CO2_KEY, Name: "Carbon Dioxide", DataType: common.DATATYPE_FLOAT, Unit: "ppm", Enable: true, Notify: false, AlarmLow: 800, AlarmHigh: 1300},
		{ID: roomDeviceMetric14ID, Key: common.METRIC_ROOM_PHOTO_KEY, Name: "Light Sensor", DataType: common.DATATYPE_INT, Unit: "mV", Enable: true, Notify: false, AlarmLow: 800, AlarmHigh: 100},
		{ID: roomDeviceMetric15ID, Key: common.METRIC_ROOM_WATERLEAK0_KEY, Name: "Pod 1 Water Leak", DataType: common.DATATYPE_INT, Unit: "mV", Enable: true, Notify: false, AlarmLow: -1, AlarmHigh: 1},
		{ID: roomDeviceMetric16ID, Key: common.METRIC_ROOM_WATERLEAK1_KEY, Name: "Pod 2 Water Leak", DataType: common.DATATYPE_INT, Unit: "mV", Enable: true, Notify: false, AlarmLow: -1, AlarmHigh: 1}})
	ventOnHours := []int{13, 14, 15, 16, 17, 18}
	ventSchedules := make([]config.Schedule, len(ventOnHours))
	for i, hour := range ventOnHours {
		key := fmt.Sprintf("room-sched-%d", i)
		id := initializer.newFarmID(farmID, key)
		ventOn := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, initializer.location)
		ventSchedules[i] = config.Schedule{ID: id, StartDate: ventOn, Frequency: common.SCHEDULE_FREQUENCY_DAILY}
	}
	roomChannel0ID := initializer.newFarmID(farmID, "room-chan-0")
	roomChannel1ID := initializer.newFarmID(farmID, "room-chan-1")
	roomChannel2ID := initializer.newFarmID(farmID, "room-chan-2")
	roomChannel3ID := initializer.newFarmID(farmID, "room-chan-3")
	roomChannel4ID := initializer.newFarmID(farmID, "room-chan-4")
	roomChannel5ID := initializer.newFarmID(farmID, "room-chan-5")
	roomChannel0ScheduleID := initializer.newFarmID(farmID, "room-chan-0-sched-1")
	roomDevice.SetChannels([]config.Channel{
		{ID: roomChannel0ID, ChannelID: common.CHANNEL_ROOM_LIGHTING_ID, Name: common.CHANNEL_ROOM_LIGHTING, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 64800, AlgorithmID: 0,
			Schedule: []config.Schedule{{ID: roomChannel0ScheduleID, StartDate: sevenPM, Frequency: common.SCHEDULE_FREQUENCY_DAILY}}},
		{ID: roomChannel1ID, ChannelID: common.CHANNEL_ROOM_AC_ID, Name: common.CHANNEL_ROOM_AC, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0},
		{ID: roomChannel2ID, ChannelID: common.CHANNEL_ROOM_HEATER_ID, Name: common.CHANNEL_ROOM_HEATER, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0},
		{ID: roomChannel3ID, ChannelID: common.CHANNEL_ROOM_DEHUEY_ID, Name: common.CHANNEL_ROOM_DEHUEY, Enable: true, Notify: true, Debounce: 10, Backoff: 0, Duration: 0, AlgorithmID: 0},
		{ID: roomChannel4ID, ChannelID: common.CHANNEL_ROOM_VENTILATION_ID, Name: common.CHANNEL_ROOM_VENTILATION, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 900, AlgorithmID: 0,
			Schedule: ventSchedules},
		{ID: roomChannel5ID, ChannelID: common.CHANNEL_ROOM_CO2_ID, Name: common.CHANNEL_ROOM_CO2, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0}})

	reservoirDevice := config.NewDevice()
	resDeviceEnableID := initializer.newFarmID(farmID, "res-enable")
	resDeviceNotifyID := initializer.newFarmID(farmID, "res-notify")
	resDeviceUriID := initializer.newFarmID(farmID, "res-uri")
	resDeviceGallonsID := initializer.newFarmID(farmID, "res-gallons")
	resChannel0ID := initializer.newFarmID(farmID, "res-chan-0")
	resChannel1ID := initializer.newFarmID(farmID, "res-chan-1")
	resChannel2ID := initializer.newFarmID(farmID, "res-chan-2")
	resChannel3ID := initializer.newFarmID(farmID, "res-chan-3")
	resChannel4ID := initializer.newFarmID(farmID, "res-chan-4")
	resDeviceWaterChangeEnableID := initializer.newFarmID(farmID, "res-waterchange-enable")
	resDeviceWaterChangeNotifyID := initializer.newFarmID(farmID, "res-waterchange-notify")
	resDeviceMetric0ID := initializer.newFarmID(farmID, "metric-0")
	resDeviceMetric1ID := initializer.newFarmID(farmID, "metric-1")
	resDeviceMetric2ID := initializer.newFarmID(farmID, "metric-2")
	resDeviceMetric3ID := initializer.newFarmID(farmID, "metric-3")
	resDeviceMetric4ID := initializer.newFarmID(farmID, "metric-4")
	resDeviceMetric5ID := initializer.newFarmID(farmID, "metric-5")
	resDeviceMetric6ID := initializer.newFarmID(farmID, "metric-6")
	resDeviceMetric7ID := initializer.newFarmID(farmID, "metric-7")
	resDeviceMetric8ID := initializer.newFarmID(farmID, "metric-8")
	resDeviceMetric9ID := initializer.newFarmID(farmID, "metric-9")
	resDeviceMetric10ID := initializer.newFarmID(farmID, "metric-10")
	resDeviceMetric11ID := initializer.newFarmID(farmID, "metric-11")
	resDeviceMetric12ID := initializer.newFarmID(farmID, "metric-12")
	resDeviceMetric13ID := initializer.newFarmID(farmID, "metric-13")
	resDeviceMetric14ID := initializer.newFarmID(farmID, "metric-14")
	reservoirDevice.SetID(initializer.newFarmID(farmID, common.CONTROLLER_TYPE_RESERVOIR))
	reservoirDevice.SetType(common.CONTROLLER_TYPE_RESERVOIR)
	reservoirDevice.SetDescription("Manages and monitors reservoir water and nutrients")
	reservoirDevice.SetConfigs([]config.DeviceConfigItem{
		{ID: resDeviceEnableID, UserID: adminUser.GetID(), DeviceID: reservoirDevice.GetID(), Key: common.CONFIG_RESERVOIR_ENABLE_KEY, Value: "true"},
		{ID: resDeviceNotifyID, UserID: adminUser.GetID(), DeviceID: reservoirDevice.GetID(), Key: common.CONFIG_RESERVOIR_NOTIFY_KEY, Value: "true"},
		{ID: resDeviceUriID, UserID: adminUser.GetID(), DeviceID: reservoirDevice.GetID(), Key: common.CONFIG_RESERVOIR_URI_KEY},
		{ID: resDeviceGallonsID, UserID: adminUser.GetID(), DeviceID: reservoirDevice.GetID(), Key: common.CONFIG_RESERVOIR_GALLONS_KEY, Value: common.DEFAULT_GALLONS},
		{ID: resDeviceWaterChangeEnableID, UserID: adminUser.GetID(), DeviceID: reservoirDevice.GetID(), Key: common.CONFIG_RESERVOIR_WATERCHANGE_ENABLE_KEY, Value: "false"},
		{ID: resDeviceWaterChangeNotifyID, UserID: adminUser.GetID(), DeviceID: reservoirDevice.GetID(), Key: common.CONFIG_RESERVOIR_WATERCHANGE_NOTIFY_KEY, Value: "false"}})
	reservoirDevice.SetMetrics([]config.Metric{
		{ID: resDeviceMetric0ID, Key: common.METRIC_RESERVOIR_MEMORY_KEY, Name: "Available System Memory", DataType: common.DATATYPE_INT, Unit: "bytes", Enable: true, Notify: true, AlarmLow: 500, AlarmHigh: 100000},
		{ID: resDeviceMetric1ID, Key: common.METRIC_RESERVOIR_TEMP_KEY, Name: "Water Temperature", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: true, AlarmLow: 61, AlarmHigh: 67},
		{ID: resDeviceMetric2ID, Key: common.METRIC_RESERVOIR_PH_KEY, Name: "pH", DataType: common.DATATYPE_FLOAT, Unit: "", Enable: true, Notify: true, AlarmLow: 5.4, AlarmHigh: 6.2},
		{ID: resDeviceMetric3ID, Key: common.METRIC_RESERVOIR_EC_KEY, Name: "Electrical Conductivity (EC)", DataType: common.DATATYPE_FLOAT, Unit: "mV", Enable: true, Notify: true, AlarmLow: 850, AlarmHigh: 1300},
		{ID: resDeviceMetric4ID, Key: common.METRIC_RESERVOIR_TDS_KEY, Name: "Total Dissolved Solids (TDS)", DataType: common.DATATYPE_FLOAT, Unit: "ppm", Enable: true, Notify: false, AlarmLow: 700, AlarmHigh: 900},
		{ID: resDeviceMetric5ID, Key: common.METRIC_RESERVOIR_ORP_KEY, Name: "Oxygen Reduction Potential (ORP)", DataType: common.DATATYPE_FLOAT, Unit: "mV", Enable: true, Notify: false, AlarmLow: 250, AlarmHigh: 375},
		{ID: resDeviceMetric6ID, Key: common.METRIC_RESERVOIR_DOMGL_KEY, Name: "Dissolved Oxygen (DO)", DataType: common.DATATYPE_FLOAT, Unit: "mg/L", Enable: true, Notify: false, AlarmLow: 5, AlarmHigh: 30},
		{ID: resDeviceMetric7ID, Key: common.METRIC_RESERVOIR_DOPER_KEY, Name: "Dissolved Oxygen (DO)", DataType: common.DATATYPE_FLOAT, Unit: "%", Enable: true, Notify: false, AlarmLow: 0, AlarmHigh: 0},
		{ID: resDeviceMetric8ID, Key: common.METRIC_RESERVOIR_SAL_KEY, Name: "Salinity", DataType: common.DATATYPE_FLOAT, Unit: "ppt", Enable: true, Notify: false, AlarmLow: 0, AlarmHigh: 0},
		{ID: resDeviceMetric9ID, Key: common.METRIC_RESERVOIR_SG_KEY, Name: "Specific Gravity", DataType: common.DATATYPE_FLOAT, Unit: "ppt", Enable: true, Notify: false, AlarmLow: 0, AlarmHigh: 0},
		{ID: resDeviceMetric10ID, Key: common.METRIC_RESERVOIR_ENVTEMP_KEY, Name: "Environment Temp", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: false, AlarmLow: 40, AlarmHigh: 80},
		{ID: resDeviceMetric11ID, Key: common.METRIC_RESERVOIR_ENVHUMIDITY_KEY, Name: "Environment Humidity", DataType: common.DATATYPE_FLOAT, Unit: "%", Enable: true, Notify: false, AlarmLow: 40, AlarmHigh: 80},
		{ID: resDeviceMetric12ID, Key: common.METRIC_RESERVOIR_ENVHEATINDEX_KEY, Name: "Environment Heat Index", DataType: common.DATATYPE_FLOAT, Unit: "%", Enable: true, Notify: false, AlarmLow: 40, AlarmHigh: 80},
		{ID: resDeviceMetric13ID, Key: common.METRIC_RESERVOIR_LOWERFLOAT_KEY, Name: "Lower Float", DataType: common.DATATYPE_INT, Unit: "", Enable: true, Notify: false, AlarmLow: -1, AlarmHigh: 1},
		{ID: resDeviceMetric14ID, Key: common.METRIC_RESERVOIR_UPPERFLOAT_KEY, Name: "Upper Float", DataType: common.DATATYPE_INT, Unit: "", Enable: true, Notify: false, AlarmLow: -1, AlarmHigh: 1}})
	// Custom auxiliary schedule
	// endDate := sevenPM.Add(time.Duration(common.HOURS_IN_A_YEAR) * time.Hour)
	// days := "SU,TU,TH,SA"
	// oneMinuteFromNow := time.Date(now.Year(), now.Month(), now.Day(), nowHr, nowMin+1, 0, 0, initializer.location)
	// Top-off schedule
	nineAM := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, initializer.location)
	ninePM := time.Date(now.Year(), now.Month(), now.Day(), 21, 0, 0, 0, initializer.location)
	weeklyID := initializer.newFarmID(farmID, "res-chan-5-sched-weekly")
	monthlyID := initializer.newFarmID(farmID, "res-chan-5-sched-monthly")
	yearlyID := initializer.newFarmID(farmID, "res-chan-5-sched-yearly")
	reservoirDevice.SetChannels([]config.Channel{
		{ID: resChannel0ID, ChannelID: common.CHANNEL_RESERVOIR_DRAIN_ID, Name: common.CHANNEL_RESERVOIR_DRAIN, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0},
		{ID: resChannel1ID, ChannelID: common.CHANNEL_RESERVOIR_CHILLER_ID, Name: common.CHANNEL_RESERVOIR_CHILLER, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0},
		{ID: resChannel2ID, ChannelID: common.CHANNEL_RESERVOIR_HEATER_ID, Name: common.CHANNEL_RESERVOIR_HEATER, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0},
		{ID: resChannel3ID, ChannelID: common.CHANNEL_RESERVOIR_POWERHEAD_ID, Name: common.CHANNEL_RESERVOIR_POWERHEAD, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0},
		{ID: resChannel4ID, ChannelID: common.CHANNEL_RESERVOIR_AUX_ID, Name: common.CHANNEL_RESERVOIR_AUX, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 60, AlgorithmID: 0}, // Schedule: []config.Schedule{
		// 	{StartDate: sevenPM, EndDate: &endDate, Frequency: common.SCHEDULE_FREQUENCY_WEEKLY, Interval: 2, Count: 5, Days: &days},
		// 	{StartDate: oneMinuteFromNow, Frequency: common.SCHEDULE_FREQUENCY_DAILY}}},
		{ChannelID: common.CHANNEL_RESERVOIR_TOPOFF_ID, Name: common.CHANNEL_RESERVOIR_TOPOFF, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 120, AlgorithmID: 0,
			Schedule: []config.Schedule{
				{ID: weeklyID, StartDate: nineAM, Frequency: common.SCHEDULE_FREQUENCY_WEEKLY},
				{ID: monthlyID, StartDate: ninePM, Frequency: common.SCHEDULE_FREQUENCY_MONTHLY},
				{ID: yearlyID, StartDate: ninePM, Frequency: common.SCHEDULE_FREQUENCY_YEARLY}}},
		{ChannelID: common.CHANNEL_RESERVOIR_FAUCET_ID, Name: common.CHANNEL_RESERVOIR_FAUCET, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0}})

	doserDevice := config.NewDevice()
	doserDeviceEnableID := initializer.newFarmID(farmID, "doser-enable")
	doserDeviceNotifyID := initializer.newFarmID(farmID, "doser-notify")
	doserDeviceUriID := initializer.newFarmID(farmID, "doser-uri")
	doserDeviceGallonsID := initializer.newFarmID(farmID, "doser-gallons")
	doserChannel0ID := initializer.newFarmID(farmID, "doser-chan-0")
	doserChannel1ID := initializer.newFarmID(farmID, "doser-chan-1")
	doserChannel2ID := initializer.newFarmID(farmID, "doser-chan-2")
	doserChannel3ID := initializer.newFarmID(farmID, "doser-chan-3")
	doserChannel4ID := initializer.newFarmID(farmID, "doser-chan-4")
	doserChannel5ID := initializer.newFarmID(farmID, "doser-chan-5")
	doserChannel6ID := initializer.newFarmID(farmID, "doser-chan-6")
	doserDevice.SetID(initializer.newFarmID(farmID, common.CONTROLLER_TYPE_DOSER))
	doserDevice.SetType(common.CONTROLLER_TYPE_DOSER)
	doserDevice.SetDescription("Nutrient dosing and expansion I/O device")
	doserDevice.SetConfigs([]config.DeviceConfigItem{
		{ID: doserDeviceEnableID, UserID: adminUser.GetID(), DeviceID: doserDevice.GetID(), Key: common.CONFIG_DOSER_ENABLE_KEY, Value: "true"},
		{ID: doserDeviceNotifyID, UserID: adminUser.GetID(), DeviceID: doserDevice.GetID(), Key: common.CONFIG_DOSER_NOTIFY_KEY, Value: "true"},
		{ID: doserDeviceUriID, UserID: adminUser.GetID(), DeviceID: doserDevice.GetID(), Key: common.CONFIG_DOSER_URI_KEY},
		{ID: doserDeviceGallonsID, UserID: adminUser.GetID(), DeviceID: doserDevice.GetID(), Key: common.CONFIG_DOSER_GALLONS_KEY, Value: common.DEFAULT_GALLONS}})
	doserDevice.SetChannels([]config.Channel{
		{ID: doserChannel0ID, ChannelID: common.CHANNEL_DOSER_PHDOWN_ID, Name: common.CHANNEL_DOSER_PHDOWN, Enable: true, Notify: true, Debounce: 0, Backoff: 10, Duration: 0, AlgorithmID: 1},
		{ID: doserChannel1ID, ChannelID: common.CHANNEL_DOSER_PHUP_ID, Name: common.CHANNEL_DOSER_PHUP, Enable: false, Notify: true, Debounce: 0, Backoff: 10, Duration: 0, AlgorithmID: 1},
		{ID: doserChannel2ID, ChannelID: common.CHANNEL_DOSER_OXIDIZER_ID, Name: common.CHANNEL_DOSER_OXIDIZER, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 2},
		{ID: doserChannel3ID, ChannelID: common.CHANNEL_DOSER_TOPOFF_ID, Name: common.CHANNEL_DOSER_TOPOFF, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 120, AlgorithmID: 0},
		{ID: doserChannel4ID, ChannelID: common.CHANNEL_DOSER_NUTE1_ID, Name: common.CHANNEL_DOSER_NUTE1, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 30, AlgorithmID: 0},
		{ID: doserChannel5ID, ChannelID: common.CHANNEL_DOSER_NUTE2_ID, Name: common.CHANNEL_DOSER_NUTE2, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 30, AlgorithmID: 0},
		{ID: doserChannel6ID, ChannelID: common.CHANNEL_DOSER_NUTE3_ID, Name: common.CHANNEL_DOSER_NUTE3, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 30, AlgorithmID: 0}})

	farm := config.NewFarm()
	farm.SetID(farmID)
	if initializer.appMode == common.MODE_CLUSTER {
		farm.StateStore = state.RAFT_STORE
		farm.ConfigStore = config.RAFT_MEMORY_STORE
		farm.DataStore = datastore.GORM_STORE
		farm.Consistency = common.CONSISTENCY_CACHED
	}
	//  else {
	// 	farm.StateStore = state.GORM_STORE
	// 	farm.ConfigStore = config.GORM_STORE
	// 	farm.DataStore = datastore.GORM_STORE
	// 	farm.Consistency = common.CONSISTENCY_LOCAL
	// }
	farm.SetDevices([]config.Device{*serverDevice, *roomDevice, *reservoirDevice, *doserDevice})
	initializer.farmDAO.Save(farm)

	permission := &config.Permission{
		UserID: adminUser.GetID(),
		RoleID: assignedRole.GetID(),
		FarmID: farm.GetID()}
	initializer.permissionDAO.Save(permission)

	// Create conditions now that metric ids have been saved to the databsae.
	// GORM has trouble managing tables  with multiple foreign keys
	farmConfig, err := initializer.farmDAO.Get(farm.GetID(), common.CONSISTENCY_LOCAL)
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
		{
			ID:         initializer.newFarmID(farmID, "room-chan-1-cond-1"),
			MetricID:   roomDevice.GetMetrics()[1].GetID(),
			Comparator: ">",
			Threshold:  85.0}})

	roomChannels[2].SetConditions([]config.Condition{
		{
			ID:         initializer.newFarmID(farmID, "room-chan-2-cond-1"),
			MetricID:   roomDevice.GetMetrics()[1].GetID(),
			Comparator: "<",
			Threshold:  70.0}})

	roomChannels[3].SetConditions([]config.Condition{
		{
			ID:         initializer.newFarmID(farmID, "room-chan-3-cond-1"),
			MetricID:   roomDevice.GetMetrics()[3].GetID(),
			Comparator: ">",
			Threshold:  55.0}})

	roomChannels[5].SetConditions([]config.Condition{
		{
			ID:         initializer.newFarmID(farmID, "room-chan-5-cond-1"),
			MetricID:   roomDevice.GetMetrics()[14].GetID(),
			Comparator: "<",
			Threshold:  1200.0}})

	reservoirChannels[1].SetConditions([]config.Condition{
		{
			ID:         initializer.newFarmID(farmID, "res-chan-1-cond-1"),
			MetricID:   reservoirDevice.GetMetrics()[2].GetID(),
			Comparator: ">",
			Threshold:  62.0}})

	reservoirChannels[2].SetConditions([]config.Condition{
		{
			ID:         initializer.newFarmID(farmID, "res-chan-2-cond-1"),
			MetricID:   reservoirDevice.GetMetrics()[2].GetID(),
			Comparator: "<",
			Threshold:  60.0}})

	doserChannels[0].SetConditions([]config.Condition{
		{
			ID:         initializer.newFarmID(farmID, "doser-chan-0-cond-1"),
			MetricID:   reservoirDevice.GetMetrics()[2].GetID(),
			Comparator: ">",
			Threshold:  6.1}})

	doserChannels[1].SetConditions([]config.Condition{
		{
			ID:         initializer.newFarmID(farmID, "doser-chan-1-cond-1"),
			MetricID:   reservoirDevice.GetMetrics()[2].GetID(),
			Comparator: "<",
			Threshold:  5.4}})

	doserChannels[2].SetConditions([]config.Condition{
		{
			ID:         initializer.newFarmID(farmID, "doser-chan-2-cond-1"),
			MetricID:   reservoirDevice.GetMetrics()[5].GetID(),
			Comparator: "<",
			Threshold:  300.0}})

	// Water change workflow
	drainStep := config.NewWorkflowStep()
	drainStep.SetID(initializer.newFarmID(farmID, "drainStep"))
	drainStep.SetDeviceID(persistedReservoir.GetID())
	drainStep.SetChannelID(reservoirDevice.GetChannels()[0].GetID()) // CHANNEL_RESERVOIR_DRAIN
	// drainStep.SetDuration(300) // seconds; 5 minutes
	// drainStep.SetWait(300)     // seconds; 5 minutes
	drainStep.SetDuration(5)
	drainStep.SetWait(10)

	fillStep := config.NewWorkflowStep()
	fillStep.SetID(initializer.newFarmID(farmID, "fillStep"))
	fillStep.SetDeviceID(persistedReservoir.GetID())
	fillStep.SetChannelID(reservoirDevice.GetChannels()[6].GetID()) // CHANNEL_RESERVOIR_FAUCET
	// fillStep.SetDuration(300) // seconds; 5 minutes
	// fillStep.SetWait(300)     // seconds; 5 minutes
	fillStep.SetDuration(5)
	fillStep.SetWait(10)

	phDownStep := config.NewWorkflowStep()
	phDownStep.SetID(initializer.newFarmID(farmID, "phDownStep"))
	phDownStep.SetDeviceID(persistedDoser.GetID())
	phDownStep.SetChannelID(doserDevice.GetChannels()[0].GetID()) // CHANNEL_DOSER_PHDOWN
	// phDownStep.SetDuration(60)  // seconds
	// phDownStep.SetWait(300)     // seconds; 5 minutes
	phDownStep.SetDuration(5)
	phDownStep.SetWait(10)

	nutePart1Step := config.NewWorkflowStep()
	nutePart1Step.SetID(initializer.newFarmID(farmID, "nutePart1Step"))
	nutePart1Step.SetDeviceID(persistedDoser.GetID())
	nutePart1Step.SetChannelID(doserDevice.GetChannels()[4].GetID()) // CHANNEL_DOSER_NUTE1
	// nutePart1Step.SetDuration(30)  // seconds
	// nutePart1Step.SetWait(300)     // seconds; 5 minutes
	nutePart1Step.SetDuration(5)
	nutePart1Step.SetWait(10)

	nutePart2Step := config.NewWorkflowStep()
	nutePart2Step.SetID(initializer.newFarmID(farmID, "nutePart2Step"))
	nutePart2Step.SetDeviceID(persistedDoser.GetID())
	nutePart2Step.SetChannelID(doserDevice.GetChannels()[5].GetID()) // CHANNEL_DOSER_NUTE2
	// nutePart2Step.SetDuration(30)  // seconds
	// nutePart2Step.SetWait(300)     // seconds; 5 minutes
	nutePart2Step.SetDuration(5)
	nutePart2Step.SetWait(10)

	nutePart3Step := config.NewWorkflowStep()
	nutePart3Step.SetID(initializer.newFarmID(farmID, "nutePart3Step"))
	nutePart3Step.SetDeviceID(persistedDoser.GetID())
	nutePart3Step.SetChannelID(doserDevice.GetChannels()[6].GetID()) // CHANNEL_DOSER_NUTE3
	// nutePart3Step.SetDuration(30)  // seconds
	// nutePart3Step.SetWait(300)     // seconds; 5 minutes
	nutePart3Step.SetDuration(5)
	nutePart3Step.SetWait(10)

	workflow1 := config.NewWorkflow()
	workflow1.SetID(initializer.newFarmID(farmID, "workflow1"))
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
