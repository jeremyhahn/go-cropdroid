package dao

import (
	"errors"
	"fmt"
	"time"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/util"
	logging "github.com/op/go-logging"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrFarmAlreadyExists = errors.New("farm already exists")
)

type Initializer interface {
	Initialize(includeFarm bool, params *common.ProvisionerParams) (*config.Farm, error)
	BuildConfig(*common.ProvisionerParams, *config.User) (*config.Farm, error)
	//DefaultFarmID() (uint64, string, error)
}

type ConfigInitializer struct {
	logger        *logging.Logger
	location      *time.Location
	algorithmDAO  AlgorithmDAO
	farmDAO       FarmDAO
	userDAO       UserDAO
	roleDAO       RoleDAO
	permissionDAO PermissionDAO
	idGenerator   util.IdGenerator
	appMode       string
}

func NewConfigInitializer(logger *logging.Logger, idGenerator util.IdGenerator,
	location *time.Location, registry Registry, appMode string) *ConfigInitializer {

	return &ConfigInitializer{
		logger:        logger,
		idGenerator:   idGenerator,
		location:      location,
		algorithmDAO:  registry.GetAlgorithmDAO(),
		farmDAO:       registry.GetFarmDAO(),
		userDAO:       registry.GetUserDAO(),
		roleDAO:       registry.GetRoleDAO(),
		permissionDAO: registry.GetPermissionDAO(),
		appMode:       appMode}
}

// Initializes a new database, including a new administrative user and default Farm.
func (initializer *ConfigInitializer) Initialize(includeFarm bool,
	params *common.ProvisionerParams) (*config.Farm, error) {

	encrypted, err := bcrypt.GenerateFromPassword([]byte(common.DEFAULT_PASSWORD), bcrypt.DefaultCost)
	if err != nil {
		initializer.logger.Fatalf("Error generating encrypted password: %s", err)
		return nil, err
	}

	adminRole := config.NewRole()
	adminRole.SetID(initializer.newID(common.DEFAULT_ROLE))
	adminRole.SetName(common.DEFAULT_ROLE)
	initializer.roleDAO.Save(adminRole)

	cultivatorRole := config.NewRole()
	cultivatorRole.SetID(initializer.newID(common.ROLE_CULTIVATOR))
	cultivatorRole.SetName(common.ROLE_CULTIVATOR)
	initializer.roleDAO.Save(cultivatorRole)

	analystRole := config.NewRole()
	analystRole.SetID(initializer.newID(common.ROLE_ANALYST))
	analystRole.SetName(common.ROLE_ANALYST)
	initializer.roleDAO.Save(analystRole)

	adminUser := config.NewUser()
	adminUser.ID = initializer.newID(common.DEFAULT_USER)
	adminUser.SetEmail(common.DEFAULT_USER)
	adminUser.SetPassword(string(encrypted))
	adminUser.SetRoles([]*config.Role{adminRole})
	initializer.userDAO.Save(adminUser)

	var farm *config.Farm
	if includeFarm {
		farm, err = initializer.BuildConfig(params, adminUser)
		if err != nil {
			return nil, err
		}
		if err := initializer.farmDAO.Save(farm); err != nil {
			return nil, err
		}
	}

	phAlgoID := initializer.newID(common.ALGORITHM_PH_KEY)
	phAlgo := &config.Algorithm{ID: phAlgoID, Name: common.ALGORITHM_PH_KEY}
	initializer.algorithmDAO.Save(phAlgo)

	// Oxidizer can be set up with a simple condition
	// oxidizerAlgoID := initializer.newID(common.ALGORITHM_ORP_KEY)
	// oxidizerAlgo := &config.Algorithm{ID: oxidizerAlgoID, Name: common.ALGORITHM_ORP_KEY}
	// initializer.algorithmDAO.Save(oxidizerAlgo)

	initializer.seedInventory()

	return farm, nil
}

func (initializer *ConfigInitializer) newID(key string) uint64 {
	return initializer.idGenerator.NewID(key)
}

func (initializer *ConfigInitializer) newFarmID(farmID uint64, key string) uint64 {
	return initializer.idGenerator.NewID(fmt.Sprintf("%d-%s", farmID, key))
}

func (initializer *ConfigInitializer) BuildConfig(params *common.ProvisionerParams, adminUser *config.User) (*config.Farm, error) {

	farmKey := fmt.Sprintf("%d-%s", params.OrganizationID, params.FarmName)
	farmID := initializer.idGenerator.NewID(farmKey)

	// Add permissions for all existing users in the organization
	users, _ := initializer.permissionDAO.GetUsers(params.OrganizationID,
		params.ConsistencyLevel)
	for _, user := range users {
		permission := config.Permission{
			UserID: user.GetID(),
			RoleID: user.GetRoles()[0].GetID(),
			FarmID: farmID}
		initializer.permissionDAO.Save(&permission)
	}

	// Add permission for user who requested the provisioning
	permission := config.Permission{
		UserID: params.UserID,
		RoleID: params.RoleID,
		FarmID: farmID}
	initializer.permissionDAO.Save(&permission)

	// TODO: This is for GORM.
	// if users == nil && adminUser != nil {
	// 	// Admin user is nil when Gossip.handleEvent
	// 	// processes a new EventProvisionRequest. Dont
	// 	// run this code for this event.
	// 	users = []*config.User{adminUser}
	// 	params.UserID = adminUser.ID
	// 	params.RoleID = adminUser.Roles[0].ID
	// }

	// Create a final list of users to assign to the farm
	adminUserID := initializer.newID(common.DEFAULT_USER)
	farmUsers := make([]*config.User, len(users))
	hasAdmin := false
	for i, user := range users {
		farmUsers[i] = user
		if user.GetID() == adminUserID {
			hasAdmin = true
		}
	}
	provisioningUser, err := initializer.userDAO.Get(params.UserID, params.ConsistencyLevel)
	if err != nil {
		return nil, err
	}
	farmUsers = append(farmUsers, provisioningUser)
	if params.UserID == adminUserID {
		hasAdmin = true
	}
	if !hasAdmin {
		adminUser, err := initializer.userDAO.Get(adminUserID, params.ConsistencyLevel)
		if err != nil {
			return nil, err
		}
		farmUsers = append(farmUsers, adminUser)
	}

	defaultTimezone := initializer.location.String()
	now := time.Now().In(initializer.location)
	sevenPM := time.Date(now.Year(), now.Month(), now.Day(), 19, 0, 0, 0, initializer.location)

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
	serverDevice.SetSettings([]*config.DeviceSetting{
		{ID: serverDeviceNameID, UserID: params.UserID, DeviceID: serverDeviceID, Key: common.CONFIG_NAME_KEY, Value: params.FarmName},
		{ID: serverDeviceIntervalID, UserID: params.UserID, DeviceID: serverDeviceID, Key: common.CONFIG_INTERVAL_KEY, Value: "60"},
		{ID: serverDeviceTimezoneID, UserID: params.UserID, DeviceID: serverDeviceID, Key: common.CONFIG_TIMEZONE_KEY, Value: defaultTimezone},
		{ID: serverDeviceModeID, UserID: params.UserID, DeviceID: serverDeviceID, Key: common.CONFIG_MODE_KEY, Value: "virtual"},
		{ID: serverDeviceSmtpEnableID, UserID: params.UserID, DeviceID: serverDeviceID, Key: common.CONFIG_SMTP_ENABLE_KEY, Value: "false"},
		{ID: serverDeviceSmtpHostID, UserID: params.UserID, DeviceID: serverDeviceID, Key: common.CONFIG_SMTP_HOST_KEY, Value: "smtp.gmail.com"},
		{ID: serverDeviceSmtpPortID, UserID: params.UserID, DeviceID: serverDeviceID, Key: common.CONFIG_SMTP_PORT_KEY, Value: "587"},
		{ID: serverDeviceSmtpUsernameID, UserID: params.UserID, DeviceID: serverDeviceID, Key: common.CONFIG_SMTP_USERNAME_KEY, Value: "myuser@gmail.com"},
		{ID: serverDeviceSmtpPasswordID, UserID: params.UserID, DeviceID: serverDeviceID, Key: common.CONFIG_SMTP_PASSWORD_KEY, Value: "$ecret!"},
		{ID: serverDeviceSmtpRecipientID, UserID: params.UserID, DeviceID: serverDeviceID, Key: common.CONFIG_SMTP_RECIPIENT_KEY, Value: "1234567890@vtext.com"}})

	roomDevice := config.NewDevice()
	roomDeviceEnableID := initializer.newFarmID(farmID, "room-enable")
	roomDeviceNotifyID := initializer.newFarmID(farmID, "room-notify")
	roomDeviceUriID := initializer.newFarmID(farmID, "room-uri")
	roomDeviceVideoID := initializer.newFarmID(farmID, "room-video")
	roomDeviceMetric0ID := initializer.newFarmID(farmID, "room-metric-0")
	roomDeviceMetric1ID := initializer.newFarmID(farmID, "room-metric-1")
	roomDeviceMetric2ID := initializer.newFarmID(farmID, "room-metric-2")
	roomDeviceMetric3ID := initializer.newFarmID(farmID, "room-metric-3")
	roomDeviceMetric4ID := initializer.newFarmID(farmID, "room-metric-4")
	roomDeviceMetric5ID := initializer.newFarmID(farmID, "room-metric-5")
	roomDeviceMetric6ID := initializer.newFarmID(farmID, "room-metric-6")
	roomDeviceMetric7ID := initializer.newFarmID(farmID, "room-metric-7")
	roomDeviceMetric8ID := initializer.newFarmID(farmID, "room-metric-8")
	roomDeviceMetric9ID := initializer.newFarmID(farmID, "room-metric-9")
	roomDeviceMetric10ID := initializer.newFarmID(farmID, "room-metric-10")
	roomDeviceMetric11ID := initializer.newFarmID(farmID, "room-metric-11")
	roomDeviceMetric12ID := initializer.newFarmID(farmID, "room-metric-12")
	roomDeviceMetric13ID := initializer.newFarmID(farmID, "room-metric-13")
	roomDeviceMetric14ID := initializer.newFarmID(farmID, "room-metric-14")
	roomDeviceMetric15ID := initializer.newFarmID(farmID, "room-metric-15")
	roomDeviceMetric16ID := initializer.newFarmID(farmID, "room-metric-16")
	roomDeviceID := initializer.newFarmID(farmID, common.CONTROLLER_TYPE_ROOM)
	roomDevice.SetID(roomDeviceID)
	roomDevice.SetType(common.CONTROLLER_TYPE_ROOM)
	roomDevice.SetDescription("Manages and monitors room climate")
	roomDevice.SetSettings([]*config.DeviceSetting{
		{ID: roomDeviceEnableID, UserID: params.UserID, DeviceID: roomDeviceID, Key: common.CONFIG_ROOM_ENABLE_KEY, Value: "true"},
		{ID: roomDeviceNotifyID, UserID: params.UserID, DeviceID: roomDeviceID, Key: common.CONFIG_ROOM_NOTIFY_KEY, Value: "true"},
		{ID: roomDeviceUriID, UserID: params.UserID, DeviceID: roomDeviceID, Key: common.CONFIG_ROOM_URI_KEY},
		{ID: roomDeviceVideoID, UserID: params.UserID, DeviceID: roomDeviceID, Key: common.CONFIG_ROOM_VIDEO_KEY}})
	roomDevice.SetMetrics([]*config.Metric{
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
	ventSchedules := make([]*config.Schedule, len(ventOnHours))
	for i, hour := range ventOnHours {
		hr := hour
		key := fmt.Sprintf("room-sched-%d", i)
		id := initializer.newFarmID(farmID, key)
		ventOn := time.Date(now.Year(), now.Month(), now.Day(), hr, 0, 0, 0, initializer.location)
		ventSchedules[i] = &config.Schedule{ID: id, StartDate: ventOn, Frequency: common.SCHEDULE_FREQUENCY_DAILY}
	}
	roomChannel0ID := initializer.newFarmID(farmID, "room-chan-0")
	roomChannel1ID := initializer.newFarmID(farmID, "room-chan-1")
	roomChannel2ID := initializer.newFarmID(farmID, "room-chan-2")
	roomChannel3ID := initializer.newFarmID(farmID, "room-chan-3")
	roomChannel4ID := initializer.newFarmID(farmID, "room-chan-4")
	roomChannel5ID := initializer.newFarmID(farmID, "room-chan-5")
	roomChannel0ScheduleID := initializer.newFarmID(farmID, "room-chan-0-sched-1")
	roomChannel1ConditionID := initializer.newFarmID(farmID, "room-chan-1-cond-1")
	roomChannel2ConditionID := initializer.newFarmID(farmID, "room-chan-2-cond-1")
	roomChannel3ConditionID := initializer.newFarmID(farmID, "room-chan-3-cond-1")
	roomDevice.SetChannels([]*config.Channel{
		{ID: roomChannel0ID, ChannelID: common.CHANNEL_ROOM_LIGHTING_ID, Name: common.CHANNEL_ROOM_LIGHTING, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 64800, AlgorithmID: 0,
			Conditions: make([]*config.Condition, 0),
			Schedule:   []*config.Schedule{{ID: roomChannel0ScheduleID, StartDate: sevenPM, Frequency: common.SCHEDULE_FREQUENCY_DAILY}}},
		{ID: roomChannel1ID, ChannelID: common.CHANNEL_ROOM_AC_ID, Name: common.CHANNEL_ROOM_AC, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0,
			Conditions: []*config.Condition{{ID: roomChannel1ConditionID, MetricID: roomDeviceMetric1ID, Comparator: ">", Threshold: 74.0}},
			Schedule:   make([]*config.Schedule, 0)},
		{ID: roomChannel2ID, ChannelID: common.CHANNEL_ROOM_HEATER_ID, Name: common.CHANNEL_ROOM_HEATER, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0,
			Conditions: []*config.Condition{{ID: roomChannel2ConditionID, MetricID: roomDeviceMetric1ID, Comparator: "<", Threshold: 68.0}},
			Schedule:   make([]*config.Schedule, 0)},
		{ID: roomChannel3ID, ChannelID: common.CHANNEL_ROOM_DEHUEY_ID, Name: common.CHANNEL_ROOM_DEHUEY, Enable: true, Notify: true, Debounce: 10, Backoff: 0, Duration: 0, AlgorithmID: 0,
			Conditions: []*config.Condition{{ID: roomChannel3ConditionID, MetricID: roomDeviceMetric2ID, Comparator: ">", Threshold: 55.0}},
			Schedule:   make([]*config.Schedule, 0)},
		{ID: roomChannel4ID, ChannelID: common.CHANNEL_ROOM_VENTILATION_ID, Name: common.CHANNEL_ROOM_VENTILATION, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 900, AlgorithmID: 0,
			Conditions: make([]*config.Condition, 0),
			Schedule:   ventSchedules},
		{ID: roomChannel5ID, ChannelID: common.CHANNEL_ROOM_CO2_ID, Name: common.CHANNEL_ROOM_CO2, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0, Conditions: make([]*config.Condition, 0), Schedule: make([]*config.Schedule, 0)}})

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
	resTopOffID := initializer.newFarmID(farmID, "res-chan-5")
	resFaucetID := initializer.newFarmID(farmID, "res-chan-6")
	resChannel1ConditionID := initializer.newFarmID(farmID, "res-chan-1-cond-1")
	resChannel2ConditionID := initializer.newFarmID(farmID, "res-chan-2-cond-1")
	resDeviceWaterChangeEnableID := initializer.newFarmID(farmID, "res-waterchange-enable")
	resDeviceWaterChangeNotifyID := initializer.newFarmID(farmID, "res-waterchange-notify")
	resDeviceMetric0ID := initializer.newFarmID(farmID, "res-metric-0")
	resDeviceMetric1ID := initializer.newFarmID(farmID, "res-metric-1")
	resDeviceMetric2ID := initializer.newFarmID(farmID, "res-metric-2")
	resDeviceMetric3ID := initializer.newFarmID(farmID, "res-metric-3")
	resDeviceMetric4ID := initializer.newFarmID(farmID, "res-metric-4")
	resDeviceMetric5ID := initializer.newFarmID(farmID, "res-metric-5")
	resDeviceMetric6ID := initializer.newFarmID(farmID, "res-metric-6")
	resDeviceMetric7ID := initializer.newFarmID(farmID, "res-metric-7")
	resDeviceMetric8ID := initializer.newFarmID(farmID, "res-metric-8")
	resDeviceMetric9ID := initializer.newFarmID(farmID, "res-metric-9")
	resDeviceMetric10ID := initializer.newFarmID(farmID, "res-metric-10")
	resDeviceMetric11ID := initializer.newFarmID(farmID, "res-metric-11")
	resDeviceMetric12ID := initializer.newFarmID(farmID, "res-metric-12")
	resDeviceMetric13ID := initializer.newFarmID(farmID, "res-metric-13")
	resDeviceMetric14ID := initializer.newFarmID(farmID, "res-metric-14")
	reservoirDeviceID := initializer.newFarmID(farmID, common.CONTROLLER_TYPE_RESERVOIR)
	reservoirDevice.SetID(reservoirDeviceID)
	reservoirDevice.SetType(common.CONTROLLER_TYPE_RESERVOIR)
	reservoirDevice.SetDescription("Manages and monitors reservoir water and nutrients")
	reservoirDevice.SetSettings([]*config.DeviceSetting{
		{ID: resDeviceEnableID, UserID: params.UserID, DeviceID: reservoirDeviceID, Key: common.CONFIG_RESERVOIR_ENABLE_KEY, Value: "true"},
		{ID: resDeviceNotifyID, UserID: params.UserID, DeviceID: reservoirDeviceID, Key: common.CONFIG_RESERVOIR_NOTIFY_KEY, Value: "true"},
		{ID: resDeviceUriID, UserID: params.UserID, DeviceID: reservoirDeviceID, Key: common.CONFIG_RESERVOIR_URI_KEY},
		{ID: resDeviceGallonsID, UserID: params.UserID, DeviceID: reservoirDeviceID, Key: common.CONFIG_RESERVOIR_GALLONS_KEY, Value: common.DEFAULT_GALLONS},
		{ID: resDeviceWaterChangeEnableID, UserID: params.UserID, DeviceID: reservoirDeviceID, Key: common.CONFIG_RESERVOIR_WATERCHANGE_ENABLE_KEY, Value: "false"},
		{ID: resDeviceWaterChangeNotifyID, UserID: params.UserID, DeviceID: reservoirDeviceID, Key: common.CONFIG_RESERVOIR_WATERCHANGE_NOTIFY_KEY, Value: "false"}})
	reservoirDevice.SetMetrics([]*config.Metric{
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
	reservoirDevice.SetChannels([]*config.Channel{
		{ID: resChannel0ID, ChannelID: common.CHANNEL_RESERVOIR_DRAIN_ID, Name: common.CHANNEL_RESERVOIR_DRAIN, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0, Conditions: make([]*config.Condition, 0), Schedule: make([]*config.Schedule, 0)},
		{ID: resChannel1ID, ChannelID: common.CHANNEL_RESERVOIR_CHILLER_ID, Name: common.CHANNEL_RESERVOIR_CHILLER, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0,
			Conditions: []*config.Condition{{ID: resChannel1ConditionID, MetricID: resDeviceMetric1ID, Comparator: ">", Threshold: 62.0}},
			Schedule:   make([]*config.Schedule, 0)},
		{ID: resChannel2ID, ChannelID: common.CHANNEL_RESERVOIR_HEATER_ID, Name: common.CHANNEL_RESERVOIR_HEATER, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0,
			Conditions: []*config.Condition{{ID: resChannel2ConditionID, MetricID: resDeviceMetric1ID, Comparator: "<", Threshold: 60.0}},
			Schedule:   make([]*config.Schedule, 0)},
		{ID: resChannel3ID, ChannelID: common.CHANNEL_RESERVOIR_POWERHEAD_ID, Name: common.CHANNEL_RESERVOIR_POWERHEAD, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0, Conditions: make([]*config.Condition, 0), Schedule: make([]*config.Schedule, 0)},
		{ID: resChannel4ID, ChannelID: common.CHANNEL_RESERVOIR_AUX_ID, Name: common.CHANNEL_RESERVOIR_AUX, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 60, AlgorithmID: 0, Conditions: make([]*config.Condition, 0), Schedule: make([]*config.Schedule, 0)}, // Schedule: []config.Schedule{
		// 	{StartDate: sevenPM, EndDate: &endDate, Frequency: common.SCHEDULE_FREQUENCY_WEEKLY, Interval: 2, Count: 5, Days: &days},
		// 	{StartDate: oneMinuteFromNow, Frequency: common.SCHEDULE_FREQUENCY_DAILY}}},
		{ID: resTopOffID, ChannelID: common.CHANNEL_RESERVOIR_TOPOFF_ID, Name: common.CHANNEL_RESERVOIR_TOPOFF, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 120, AlgorithmID: 0,
			Conditions: make([]*config.Condition, 0),
			Schedule: []*config.Schedule{
				{ID: weeklyID, StartDate: nineAM, Frequency: common.SCHEDULE_FREQUENCY_WEEKLY},
				{ID: monthlyID, StartDate: ninePM, Frequency: common.SCHEDULE_FREQUENCY_MONTHLY},
				{ID: yearlyID, StartDate: ninePM, Frequency: common.SCHEDULE_FREQUENCY_YEARLY}}},
		{ID: resFaucetID, ChannelID: common.CHANNEL_RESERVOIR_FAUCET_ID, Name: common.CHANNEL_RESERVOIR_FAUCET, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0, Conditions: make([]*config.Condition, 0), Schedule: make([]*config.Schedule, 0)}})

	doserDevice := config.NewDevice()
	doserDeviceEnableID := initializer.newFarmID(farmID, "doser-enable")
	doserDeviceNotifyID := initializer.newFarmID(farmID, "doser-notify")
	doserDeviceUriID := initializer.newFarmID(farmID, "doser-uri")
	doserDeviceGallonsID := initializer.newFarmID(farmID, "doser-gallons")
	doserChannel0ID := initializer.newFarmID(farmID, "doser-chan-0")
	doserChannel0ConditionID := initializer.newFarmID(farmID, "doser-chan-0-cond-1")
	doserChannel1ConditionID := initializer.newFarmID(farmID, "doser-chan-1-cond-1")
	doserChannel2ConditionID := initializer.newFarmID(farmID, "doser-chan-2-cond-1")
	doserChannel1ID := initializer.newFarmID(farmID, "doser-chan-1")
	doserChannel2ID := initializer.newFarmID(farmID, "doser-chan-2")
	doserChannel3ID := initializer.newFarmID(farmID, "doser-chan-3")
	doserChannel4ID := initializer.newFarmID(farmID, "doser-chan-4")
	doserChannel5ID := initializer.newFarmID(farmID, "doser-chan-5")
	doserChannel6ID := initializer.newFarmID(farmID, "doser-chan-6")
	doserDeviceID := initializer.newFarmID(farmID, common.CONTROLLER_TYPE_DOSER)
	doserDevice.SetID(doserDeviceID)
	doserDevice.SetType(common.CONTROLLER_TYPE_DOSER)
	doserDevice.SetDescription("Nutrient dosing and expansion I/O device")
	doserDevice.SetSettings([]*config.DeviceSetting{
		{ID: doserDeviceEnableID, UserID: params.UserID, DeviceID: doserDeviceID, Key: common.CONFIG_DOSER_ENABLE_KEY, Value: "true"},
		{ID: doserDeviceNotifyID, UserID: params.UserID, DeviceID: doserDeviceID, Key: common.CONFIG_DOSER_NOTIFY_KEY, Value: "true"},
		{ID: doserDeviceUriID, UserID: params.UserID, DeviceID: doserDeviceID, Key: common.CONFIG_DOSER_URI_KEY},
		{ID: doserDeviceGallonsID, UserID: params.UserID, DeviceID: doserDeviceID, Key: common.CONFIG_DOSER_GALLONS_KEY, Value: common.DEFAULT_GALLONS}})
	doserDevice.SetChannels([]*config.Channel{
		{ID: doserChannel0ID, ChannelID: common.CHANNEL_DOSER_PHDOWN_ID, Name: common.CHANNEL_DOSER_PHDOWN, Enable: true, Notify: true, Debounce: 0, Backoff: 10, Duration: 0, AlgorithmID: 1,
			Conditions: []*config.Condition{{ID: doserChannel0ConditionID, MetricID: resDeviceMetric2ID, Comparator: ">", Threshold: 6.1}},
			Schedule:   make([]*config.Schedule, 0)},
		{ID: doserChannel1ID, ChannelID: common.CHANNEL_DOSER_PHUP_ID, Name: common.CHANNEL_DOSER_PHUP, Enable: false, Notify: true, Debounce: 0, Backoff: 10, Duration: 0, AlgorithmID: 1,
			Conditions: []*config.Condition{{ID: doserChannel1ConditionID, MetricID: resDeviceMetric2ID, Comparator: "<", Threshold: 5.4}},
			Schedule:   make([]*config.Schedule, 0)},
		{ID: doserChannel2ID, ChannelID: common.CHANNEL_DOSER_OXIDIZER_ID, Name: common.CHANNEL_DOSER_OXIDIZER, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0,
			Conditions: []*config.Condition{{ID: doserChannel2ConditionID, MetricID: resDeviceMetric5ID, Comparator: "<", Threshold: 300.0}},
			Schedule:   make([]*config.Schedule, 0)},
		{ID: doserChannel3ID, ChannelID: common.CHANNEL_DOSER_TOPOFF_ID, Name: common.CHANNEL_DOSER_TOPOFF, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 120, AlgorithmID: 0, Conditions: make([]*config.Condition, 0), Schedule: make([]*config.Schedule, 0)},
		{ID: doserChannel4ID, ChannelID: common.CHANNEL_DOSER_NUTE1_ID, Name: common.CHANNEL_DOSER_NUTE1, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 30, AlgorithmID: 0, Conditions: make([]*config.Condition, 0), Schedule: make([]*config.Schedule, 0)},
		{ID: doserChannel5ID, ChannelID: common.CHANNEL_DOSER_NUTE2_ID, Name: common.CHANNEL_DOSER_NUTE2, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 30, AlgorithmID: 0, Conditions: make([]*config.Condition, 0), Schedule: make([]*config.Schedule, 0)},
		{ID: doserChannel6ID, ChannelID: common.CHANNEL_DOSER_NUTE3_ID, Name: common.CHANNEL_DOSER_NUTE3, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 30, AlgorithmID: 0, Conditions: make([]*config.Condition, 0), Schedule: make([]*config.Schedule, 0)}})

	farm := config.NewFarm()
	farm.SetOrganizationID(params.OrganizationID)
	farm.SetID(farmID)
	farm.SetInterval(60)
	farm.SetTimezone(initializer.location.String())
	farm.SetName(params.FarmName)
	farm.SetMode(common.CONFIG_MODE_VIRTUAL)
	farm.SetConfigStore(params.ConfigStoreType)
	farm.SetStateStore(int(params.StateStoreType))
	farm.SetDataStore(params.DataStoreType)
	farm.SetConsistencyLevel(params.ConsistencyLevel)
	farm.SetDevices([]*config.Device{serverDevice,
		roomDevice, reservoirDevice, doserDevice})
	farm.SetUsers(farmUsers)
	//initializer.farmDAO.Save(farm)

	// Create conditions now that metric ids have been saved to the databsae.
	// GORM has trouble managing tables  with multiple foreign keys
	// farm, err = initializer.farmDAO.Get(farm.GetID(), common.CONSISTENCY_LOCAL)
	// if err != nil {
	// 	return nil, err
	// }

	// Water change workflow
	drainStep := config.NewWorkflowStep()
	drainStep.SetID(initializer.newFarmID(farmID, "drainStep"))
	drainStep.SetDeviceID(reservoirDeviceID)
	drainStep.SetChannelID(reservoirDevice.GetChannels()[0].GetID()) // CHANNEL_RESERVOIR_DRAIN
	// drainStep.SetDuration(300) // seconds; 5 minutes
	// drainStep.SetWait(300)     // seconds; 5 minutes
	drainStep.SetDuration(5)
	drainStep.SetWait(10)

	fillStep := config.NewWorkflowStep()
	fillStep.SetID(initializer.newFarmID(farmID, "fillStep"))
	fillStep.SetDeviceID(reservoirDeviceID)
	fillStep.SetChannelID(reservoirDevice.GetChannels()[6].GetID()) // CHANNEL_RESERVOIR_FAUCET
	// fillStep.SetDuration(300) // seconds; 5 minutes
	// fillStep.SetWait(300)     // seconds; 5 minutes
	fillStep.SetDuration(5)
	fillStep.SetWait(10)

	phDownStep := config.NewWorkflowStep()
	phDownStep.SetID(initializer.newFarmID(farmID, "phDownStep"))
	phDownStep.SetDeviceID(doserDeviceID)
	phDownStep.SetChannelID(doserDevice.GetChannels()[0].GetID()) // CHANNEL_DOSER_PHDOWN
	// phDownStep.SetDuration(60)  // seconds
	// phDownStep.SetWait(300)     // seconds; 5 minutes
	phDownStep.SetDuration(5)
	phDownStep.SetWait(10)

	nutePart1Step := config.NewWorkflowStep()
	nutePart1Step.SetID(initializer.newFarmID(farmID, "nutePart1Step"))
	nutePart1Step.SetDeviceID(doserDeviceID)
	nutePart1Step.SetChannelID(doserDevice.GetChannels()[4].GetID()) // CHANNEL_DOSER_NUTE1
	// nutePart1Step.SetDuration(30)  // seconds
	// nutePart1Step.SetWait(300)     // seconds; 5 minutes
	nutePart1Step.SetDuration(5)
	nutePart1Step.SetWait(10)

	nutePart2Step := config.NewWorkflowStep()
	nutePart2Step.SetID(initializer.newFarmID(farmID, "nutePart2Step"))
	nutePart2Step.SetDeviceID(doserDeviceID)
	nutePart2Step.SetChannelID(doserDevice.GetChannels()[5].GetID()) // CHANNEL_DOSER_NUTE2
	// nutePart2Step.SetDuration(30)  // seconds
	// nutePart2Step.SetWait(300)     // seconds; 5 minutes
	nutePart2Step.SetDuration(5)
	nutePart2Step.SetWait(10)

	nutePart3Step := config.NewWorkflowStep()
	nutePart3Step.SetID(initializer.newFarmID(farmID, "nutePart3Step"))
	nutePart3Step.SetDeviceID(doserDeviceID)
	nutePart3Step.SetChannelID(doserDevice.GetChannels()[6].GetID()) // CHANNEL_DOSER_NUTE3
	// nutePart3Step.SetDuration(30)  // seconds
	// nutePart3Step.SetWait(300)     // seconds; 5 minutes
	nutePart3Step.SetDuration(5)
	nutePart3Step.SetWait(10)

	workflow1 := config.NewWorkflow()
	workflow1.SetID(initializer.newFarmID(farmID, "workflow1"))
	workflow1.SetName("Automated Water Changes")
	workflow1.SetSteps([]*config.WorkflowStep{
		drainStep,
		fillStep,
		phDownStep,
		nutePart1Step,
		nutePart2Step,
		nutePart3Step})

	farm.SetWorkflows([]*config.Workflow{workflow1})

	return farm, nil
}

// Initializes the initial product inventory
func (initializer *ConfigInitializer) seedInventory() {
	// initializer.db.Create(&entity.InventoryType{
	// 	Name:             "pH Probe",
	// 	Description:      "+/- 0.1 resolution",
	// 	Image:            "https://www.atlas-scientific.com/_images/large_images/probes/c-ph-box.png",
	// 	LifeExpectancy:   47336400, // 1.5 yrs
	// 	MaintenanceCycle: 31557600, // 1 yr
	// 	ProductPage:      "https://www.atlas-scientific.com/product_pages/probes/c-ph-probe.html"})
	// initializer.db.Create(&entity.InventoryType{
	// 	Name:             "Oxygen Reduction Potential Probe (ORP)",
	// 	Description:      "+/– 1.1mV accuracy",
	// 	Image:            "https://www.atlas-scientific.com/_images/large_images/probes/c-orp-box.png",
	// 	MaintenanceCycle: 31557600, // 1 yr
	// 	LifeExpectancy:   47336400, // 1.5 yrs
	// 	ProductPage:      "https://www.atlas-scientific.com/product_pages/probes/c-orp-probe.html"})
	// initializer.db.Create(&entity.InventoryType{
	// 	Name:             "Conductivity Probe (EC/TDS)",
	// 	Description:      "+/ – 2% accuracy",
	// 	Image:            "https://www.atlas-scientific.com/_images/large_images/probes/ec_k0-1.png",
	// 	LifeExpectancy:   315576000, // 10 yrs
	// 	MaintenanceCycle: 315576000,
	// 	ProductPage:      "https://www.atlas-scientific.com/product_pages/probes/ec_k0-1.html"})
	// initializer.db.Create(&entity.InventoryType{
	// 	Name:             "Dissolved Oxygen Probe (DO)",
	// 	Description:      "+/ – 2% accuracy",
	// 	Image:            "https://www.atlas-scientific.com/_images/large_images/probes/do_box.png",
	// 	LifeExpectancy:   157788000, // 5 yrs
	// 	MaintenanceCycle: 47336400,  // 18 months
	// 	ProductPage:      "https://www.atlas-scientific.com/product_pages/probes/do_probe.html"})
	// initializer.db.Create(&entity.InventoryType{
	// 	Name:           "NDIR Carbon Dioxide Sensor (CO2)",
	// 	Description:    "+/- 3%  +/- 30 ppm",
	// 	Image:          "https://dqzrr9k4bjpzk.cloudfront.net/images/805759/897784873.jpg",
	// 	LifeExpectancy: 173566800, // 5.5 yrs
	// 	ProductPage:    "https://www.atlas-scientific.com/product_pages/probes/ezo-co2.html"})
	// initializer.db.Create(&entity.InventoryType{
	// 	Name:           "pH Calibration Fluids",
	// 	Description:    "3x NIST reference calibrated 500ml bottles (4.0, 7.0, 10.0)",
	// 	Image:          "https://images-na.ssl-images-amazon.com/images/I/81HsYv4kNTL._SL1500_.jpg",
	// 	LifeExpectancy: 31557600, // 1 yr
	// 	ProductPage:    "https://www.amazon.com/Biopharm-Calibration-Solution-Standards-Traceable/dp/B01E7U873K"})
	// initializer.db.Create(&entity.InventoryType{
	// 	Name:           "pH UP (BASE)",
	// 	Description:    "General Hydroponics pH Up Liquid Fertilizer, 1-Gallon",
	// 	Image:          "https://images-na.ssl-images-amazon.com/images/I/61mHEr-obpL._AC_SL1200_.jpg",
	// 	LifeExpectancy: 31557600, // 1 yr
	// 	ProductPage:    "https://www.amazon.com/General-Hydroponics-Liquid-Fertilizer-1-Gallon/dp/B000FF5V90"})
	// initializer.db.Create(&entity.InventoryType{
	// 	Name:           "pH DOWN (ACID)",
	// 	Description:    "General Hydroponics HGC722125 Liquid Premium Buffering for pH Stability, 1-Gallon, Orange",
	// 	Image:          "https://images-na.ssl-images-amazon.com/images/I/71E-fJ-tlsL._AC_SL1500_.jpg",
	// 	LifeExpectancy: 31557600, // 1 yr
	// 	ProductPage:    "https://www.amazon.com/General-Hydroponics-Liquid-Fertilizer-1-Gallon/dp/B000FG0F9U"})
}
