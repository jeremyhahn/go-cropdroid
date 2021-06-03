package gorm

import (
	"time"

	"github.com/jeremyhahn/cropdroid/common"
	"github.com/jeremyhahn/cropdroid/config"
	"github.com/jeremyhahn/cropdroid/datastore/gorm/entity"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
	"golang.org/x/crypto/bcrypt"
)

type ClusterInitializer struct {
	logger   *logging.Logger
	db       *gorm.DB
	location *time.Location
	DatabaseInitializer
}

func NewGormClusterInitializer(logger *logging.Logger, db *gorm.DB, location *time.Location) DatabaseInitializer {
	return &ClusterInitializer{
		logger:   logger,
		db:       db,
		location: location}
}

func (initializer *ClusterInitializer) Initialize() error {

	db := initializer.db

	db.LogMode(true)

	db.AutoMigrate(&config.Permission{})
	db.AutoMigrate(&config.User{})
	db.AutoMigrate(&config.Role{})
	db.AutoMigrate(&config.Controller{})
	db.AutoMigrate(&config.ControllerConfigItem{})
	db.AutoMigrate(&config.Metric{})
	db.AutoMigrate(&config.Channel{})
	db.AutoMigrate(&config.Condition{})
	db.AutoMigrate(&config.Algorithm{})
	db.AutoMigrate(&entity.EventLog{})
	db.AutoMigrate(&config.Schedule{})
	db.AutoMigrate(&config.Farm{})
	db.AutoMigrate(&config.Organization{})
	db.AutoMigrate(&config.License{})
	db.AutoMigrate(&entity.InventoryType{})
	db.AutoMigrate(&entity.Inventory{})

	defaultTimezone := initializer.location.String()
	now := time.Now().In(initializer.location)
	nowHr, nowMin, _ := now.Clock()
	sevenPM := time.Date(now.Year(), now.Month(), now.Day(), 19, 0, 0, 0, initializer.location)

	farmDAO := NewFarmDAO(initializer.logger, initializer.db)

	encrypted, err := bcrypt.GenerateFromPassword([]byte(common.DEFAULT_PASSWORD), bcrypt.DefaultCost)
	if err != nil {
		initializer.logger.Fatalf("Error generating encrypted password: %s", err)
		return err
	}

	adminUser := config.NewUser()
	adminUser.SetEmail(common.DEFAULT_USER)
	adminUser.SetPassword(string(encrypted))

	db.Create(adminUser)

	adminRole := config.NewRole()
	adminRole.SetName(common.DEFAULT_ROLE)

	db.Create(adminRole)
	db.Create(&config.Role{Name: "cultivator"})
	db.Create(&config.Role{Name: "analyst"})

	// common.SERVER_CONTROLLER_ID needs to match the ID of this server controller!!
	serverController := config.NewController()
	serverController.SetType(common.CONTROLLER_TYPE_SERVER)
	serverController.SetDescription("Provides monitoring, real-time notifications, and web services")
	serverController.SetConfigs([]config.ControllerConfigItem{
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: serverController.GetID(), Key: common.CONFIG_NAME_KEY, Value: "First Room"},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: serverController.GetID(), Key: common.CONFIG_INTERVAL_KEY, Value: "60"},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: serverController.GetID(), Key: common.CONFIG_TIMEZONE_KEY, Value: defaultTimezone},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: serverController.GetID(), Key: common.CONFIG_MODE_KEY, Value: "virtual"},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: serverController.GetID(), Key: common.CONFIG_SMTP_ENABLE_KEY, Value: "false"},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: serverController.GetID(), Key: common.CONFIG_SMTP_HOST_KEY, Value: "smtp.gmail.com"},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: serverController.GetID(), Key: common.CONFIG_SMTP_PORT_KEY, Value: "587"},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: serverController.GetID(), Key: common.CONFIG_SMTP_USERNAME_KEY, Value: "myuser@gmail.com"},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: serverController.GetID(), Key: common.CONFIG_SMTP_PASSWORD_KEY, Value: "$ecret!"},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: serverController.GetID(), Key: common.CONFIG_SMTP_RECIPIENT_KEY, Value: "1234567890@vtext.com"}})

	roomController := config.NewController()
	roomController.SetType(common.CONTROLLER_TYPE_ROOM)
	roomController.SetDescription("Manages and monitors room climate")
	roomController.SetConfigs([]config.ControllerConfigItem{
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: roomController.GetID(), Key: CONFIG_ROOM_ENABLE_KEY, Value: "true"},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: roomController.GetID(), Key: CONFIG_ROOM_NOTIFY_KEY, Value: "true"},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: roomController.GetID(), Key: CONFIG_ROOM_URI_KEY},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: roomController.GetID(), Key: CONFIG_ROOM_VIDEO_KEY}})
	roomController.SetMetrics([]config.Metric{
		config.Metric{Key: METRIC_ROOM_MEMORY_KEY, Name: "Available System Memory", DataType: common.DATATYPE_INT, Unit: "bytes", Enable: true, Notify: true, AlarmLow: 500, AlarmHigh: 100000},
		config.Metric{Key: METRIC_ROOM_TEMPF0_KEY, Name: "Ceiling Air Temperature", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: true, AlarmLow: 71, AlarmHigh: 85},
		config.Metric{Key: METRIC_ROOM_HUMIDITY0_KEY, Name: "Ceiling Humidity", DataType: common.DATATYPE_FLOAT, Unit: "%", Enable: true, Notify: true, AlarmLow: 40, AlarmHigh: 70},
		config.Metric{Key: METRIC_ROOM_HEATINDEX0_KEY, Name: "Ceiling Heat Index", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: true, AlarmLow: 71, AlarmHigh: 85},
		config.Metric{Key: METRIC_ROOM_TEMPF1_KEY, Name: "Canopy Air Temperature", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: false, Notify: true, AlarmLow: 71, AlarmHigh: 85},
		config.Metric{Key: METRIC_ROOM_HUMIDITY1_KEY, Name: "Canopy Humidity", DataType: common.DATATYPE_FLOAT, Unit: "%", Enable: true, Notify: false, AlarmLow: 71, AlarmHigh: 85},
		config.Metric{Key: METRIC_ROOM_HEATINDEX1_KEY, Name: "Canopy Heat Index", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: false, AlarmLow: 71, AlarmHigh: 85},
		config.Metric{Key: METRIC_ROOM_TEMPF2_KEY, Name: "Floor Air Temperature", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: false, AlarmLow: 71, AlarmHigh: 85},
		config.Metric{Key: METRIC_ROOM_HUMIDITY2_KEY, Name: "Floor Humidity", DataType: common.DATATYPE_FLOAT, Unit: "%", Enable: true, Notify: false, AlarmLow: 71, AlarmHigh: 85},
		config.Metric{Key: METRIC_ROOM_HEATINDEX2_KEY, Name: "Floor Heat Index", DataType: common.DATATYPE_FLOAT, Unit: "bytes", Enable: true, Notify: false, AlarmLow: 71, AlarmHigh: 85},
		config.Metric{Key: METRIC_ROOM_WATERTEMP0_KEY, Name: "Pod 1 Water Temperature", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: true, AlarmLow: 61, AlarmHigh: 67},
		config.Metric{Key: METRIC_ROOM_WATERTEMP1_KEY, Name: "Pod 2 Water Temperature", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: true, AlarmLow: 61, AlarmHigh: 67},
		config.Metric{Key: METRIC_ROOM_VPD_KEY, Name: "Vapor Pressure Deficit", DataType: common.DATATYPE_FLOAT, Unit: "", Enable: true, Notify: false, AlarmLow: -2, AlarmHigh: 2},
		config.Metric{Key: METRIC_ROOM_CO2_KEY, Name: "Carbon Dioxide", DataType: common.DATATYPE_FLOAT, Unit: "ppm", Enable: true, Notify: false, AlarmLow: 800, AlarmHigh: 1300},
		config.Metric{Key: METRIC_ROOM_PHOTO_KEY, Name: "Light Sensor", DataType: common.DATATYPE_INT, Unit: "mV", Enable: true, Notify: false, AlarmLow: 800, AlarmHigh: 100},
		config.Metric{Key: METRIC_ROOM_WATERLEAK0_KEY, Name: "Pod 1 Water Leak", DataType: common.DATATYPE_INT, Unit: "mV", Enable: true, Notify: false, AlarmLow: -1, AlarmHigh: 1},
		config.Metric{Key: METRIC_ROOM_WATERLEAK1_KEY, Name: "Pod 2 Water Leak", DataType: common.DATATYPE_INT, Unit: "mV", Enable: true, Notify: false, AlarmLow: -1, AlarmHigh: 1}})
	ventOnHours := []int{13, 14, 15, 16, 17, 18}
	ventSchedules := make([]config.Schedule, len(ventOnHours))
	for i, hour := range ventOnHours {
		ventOn := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, initializer.location)
		ventSchedules[i] = config.Schedule{StartDate: ventOn, Frequency: common.SCHEDULE_FREQUENCY_DAILY}
	}
	roomController.SetChannels([]config.Channel{
		config.Channel{ChannelID: 0, Name: CHANNEL_ROOM_LIGHTING, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 64800, AlgorithmID: 0,
			Schedule: []config.Schedule{config.Schedule{StartDate: sevenPM, Frequency: common.SCHEDULE_FREQUENCY_DAILY}}},
		config.Channel{ChannelID: 1, Name: CHANNEL_ROOM_AC, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0},
		config.Channel{ChannelID: 2, Name: CHANNEL_ROOM_HEATER, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0},
		config.Channel{ChannelID: 3, Name: CHANNEL_ROOM_DEHUEY, Enable: true, Notify: true, Debounce: 10, Backoff: 0, Duration: 0, AlgorithmID: 0},
		config.Channel{ChannelID: 4, Name: CHANNEL_ROOM_VENTILATION, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 900, AlgorithmID: 0,
			Schedule: ventSchedules},
		config.Channel{ChannelID: 5, Name: CHANNEL_ROOM_CO2, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0}})

	reservoirController := config.NewController()
	reservoirController.SetType(common.CONTROLLER_TYPE_RESERVOIR)
	reservoirController.SetDescription("Manages and monitors reservoir water and nutrients")
	reservoirController.SetConfigs([]config.ControllerConfigItem{
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: reservoirController.GetID(), Key: CONFIG_RESERVOIR_ENABLE_KEY, Value: "true"},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: reservoirController.GetID(), Key: CONFIG_RESERVOIR_NOTIFY_KEY, Value: "true"},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: reservoirController.GetID(), Key: CONFIG_RESERVOIR_URI_KEY},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: reservoirController.GetID(), Key: CONFIG_RESERVOIR_GALLONS_KEY, Value: DEFAULT_GALLONS},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: reservoirController.GetID(), Key: CONFIG_RESERVOIR_WATERCHANGE_ENABLE_KEY, Value: "false"},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: reservoirController.GetID(), Key: CONFIG_RESERVOIR_WATERCHANGE_NOTIFY_KEY, Value: "false"}})
	reservoirController.SetMetrics([]config.Metric{
		config.Metric{Key: METRIC_RESERVOIR_MEMORY_KEY, Name: "Available System Memory", DataType: common.DATATYPE_INT, Unit: "bytes", Enable: true, Notify: true, AlarmLow: 500, AlarmHigh: 100000},
		config.Metric{Key: METRIC_RESERVOIR_TEMP_KEY, Name: "Water Temperature", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: true, AlarmLow: 61, AlarmHigh: 67},
		config.Metric{Key: METRIC_RESERVOIR_PH_KEY, Name: "pH", DataType: common.DATATYPE_FLOAT, Unit: "", Enable: true, Notify: true, AlarmLow: 5.4, AlarmHigh: 6.2},
		config.Metric{Key: METRIC_RESERVOIR_EC_KEY, Name: "Electrical Conductivity (EC)", DataType: common.DATATYPE_FLOAT, Unit: "mV", Enable: true, Notify: true, AlarmLow: 850, AlarmHigh: 1300},
		config.Metric{Key: METRIC_RESERVOIR_TDS_KEY, Name: "Total Dissolved Solids (TDS)", DataType: common.DATATYPE_FLOAT, Unit: "ppm", Enable: true, Notify: false, AlarmLow: 700, AlarmHigh: 900},
		config.Metric{Key: METRIC_RESERVOIR_ORP_KEY, Name: "Oxygen Reduction Potential (ORP)", DataType: common.DATATYPE_FLOAT, Unit: "mV", Enable: true, Notify: false, AlarmLow: 250, AlarmHigh: 375},
		config.Metric{Key: METRIC_RESERVOIR_DOMGL_KEY, Name: "Dissolved Oxygen (DO)", DataType: common.DATATYPE_FLOAT, Unit: "mg/L", Enable: true, Notify: false, AlarmLow: 5, AlarmHigh: 30},
		config.Metric{Key: METRIC_RESERVOIR_DOPER_KEY, Name: "Dissolved Oxygen (DO)", DataType: common.DATATYPE_FLOAT, Unit: "%", Enable: true, Notify: false, AlarmLow: 0, AlarmHigh: 0},
		config.Metric{Key: METRIC_RESERVOIR_SAL_KEY, Name: "Salinity", DataType: common.DATATYPE_FLOAT, Unit: "ppt", Enable: true, Notify: false, AlarmLow: 0, AlarmHigh: 0},
		config.Metric{Key: METRIC_RESERVOIR_SG_KEY, Name: "Specific Gravity", DataType: common.DATATYPE_FLOAT, Unit: "ppt", Enable: true, Notify: false, AlarmLow: 0, AlarmHigh: 0},
		config.Metric{Key: METRIC_RESERVOIR_ENVTEMP_KEY, Name: "Environment Temp", DataType: common.DATATYPE_FLOAT, Unit: "°", Enable: true, Notify: false, AlarmLow: 40, AlarmHigh: 80},
		config.Metric{Key: METRIC_RESERVOIR_ENVHUMIDITY_KEY, Name: "Environment Humidity", DataType: common.DATATYPE_FLOAT, Unit: "%", Enable: true, Notify: false, AlarmLow: 40, AlarmHigh: 80},
		config.Metric{Key: METRIC_RESERVOIR_ENVHEATINDEX_KEY, Name: "Environment Heat Index", DataType: common.DATATYPE_FLOAT, Unit: "%", Enable: true, Notify: false, AlarmLow: 40, AlarmHigh: 80},
		config.Metric{Key: METRIC_RESERVOIR_LOWERFLOAT_KEY, Name: "Lower Float", DataType: common.DATATYPE_INT, Unit: "", Enable: true, Notify: false, AlarmLow: -1, AlarmHigh: 1},
		config.Metric{Key: METRIC_RESERVOIR_UPPERFLOAT_KEY, Name: "Upper Float", DataType: common.DATATYPE_INT, Unit: "", Enable: true, Notify: false, AlarmLow: -1, AlarmHigh: 1}})
	// Custom auxiliary schedule
	endDate := sevenPM.Add(time.Duration(HOURS_IN_A_YEAR) * time.Hour)
	days := "SU,TU,TH,SA"
	oneMinuteFromNow := time.Date(now.Year(), now.Month(), now.Day(), nowHr, nowMin+1, 0, 0, initializer.location)
	// Top-off schedule
	nineAM := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, initializer.location)
	ninePM := time.Date(now.Year(), now.Month(), now.Day(), 21, 0, 0, 0, initializer.location)
	reservoirController.SetChannels([]config.Channel{
		config.Channel{ChannelID: 0, Name: CHANNEL_RESERVOIR_DRAIN, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0},
		config.Channel{ChannelID: 1, Name: CHANNEL_RESERVOIR_CHILLER, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0},
		config.Channel{ChannelID: 2, Name: CHANNEL_RESERVOIR_HEATER, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0},
		config.Channel{ChannelID: 3, Name: CHANNEL_RESERVOIR_POWERHEAD, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0},
		config.Channel{ChannelID: 4, Name: CHANNEL_RESERVOIR_AUX, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 60, AlgorithmID: 0,
			Schedule: []config.Schedule{
				config.Schedule{StartDate: sevenPM, EndDate: &endDate, Frequency: common.SCHEDULE_FREQUENCY_WEEKLY, Interval: 2, Count: 5, Days: &days},
				config.Schedule{StartDate: oneMinuteFromNow, Frequency: common.SCHEDULE_FREQUENCY_DAILY}}},
		config.Channel{ChannelID: 5, Name: CHANNEL_RESERVOIR_TOPOFF, Enable: true, Notify: true, Debounce: 0, Backoff: 0, Duration: 120, AlgorithmID: 0,
			Schedule: []config.Schedule{
				config.Schedule{StartDate: nineAM, Frequency: common.SCHEDULE_FREQUENCY_WEEKLY},
				config.Schedule{StartDate: ninePM, Frequency: common.SCHEDULE_FREQUENCY_MONTHLY},
				config.Schedule{StartDate: ninePM, Frequency: common.SCHEDULE_FREQUENCY_YEARLY}}},
		config.Channel{ChannelID: 6, Name: CHANNEL_RESERVOIR_FAUCET, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 0}})

	doserController := config.NewController()
	doserController.SetType(common.CONTROLLER_TYPE_DOSER)
	doserController.SetDescription("Nutrient dosing and expansion I/O controller")
	doserController.SetConfigs([]config.ControllerConfigItem{
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: doserController.GetID(), Key: CONFIG_DOSER_ENABLE_KEY, Value: "true"},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: doserController.GetID(), Key: CONFIG_DOSER_NOTIFY_KEY, Value: "true"},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: doserController.GetID(), Key: CONFIG_DOSER_URI_KEY},
		config.ControllerConfigItem{UserID: adminUser.GetID(), ControllerID: doserController.GetID(), Key: CONFIG_DOSER_GALLONS_KEY, Value: DEFAULT_GALLONS}})
	doserController.SetChannels([]config.Channel{
		config.Channel{ChannelID: 0, Name: CHANNEL_DOSER_PHDOWN, Enable: true, Notify: true, Debounce: 0, Backoff: 10, Duration: 0, AlgorithmID: 1},
		config.Channel{ChannelID: 1, Name: CHANNEL_DOSER_PHUP, Enable: false, Notify: true, Debounce: 0, Backoff: 10, Duration: 0, AlgorithmID: 1},
		config.Channel{ChannelID: 2, Name: CHANNEL_DOSER_OXIDIZER, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 0, AlgorithmID: 2},
		config.Channel{ChannelID: 3, Name: CHANNEL_DOSER_TOPOFF, Enable: false, Notify: true, Debounce: 0, Backoff: 0, Duration: 120, AlgorithmID: 0}})

	farm := config.NewFarm()
	farm.SetControllers([]config.Controller{*serverController, *roomController, *reservoirController, *doserController})
	farmDAO.Create(farm)

	db.Create(&config.Permission{
		UserID: adminUser.GetID(),
		RoleID: adminRole.GetID(),
		FarmID: farm.GetID()})

	db.Create(&config.Algorithm{Name: "pH"})
	db.Create(&config.Algorithm{Name: "Oxidizer"})

	// Create conditions now that metric ids have been saved to the databsae. GORM has trouble managing tables
	// with multiple foreign keys
	farmConfig, err := farmDAO.Get(farm.GetID())
	if err != nil {
		return err
	}

	controllers := farmConfig.GetControllers()
	roomChannels := controllers[1].GetChannels()
	reservoirChannels := controllers[2].GetChannels()
	doserChannels := controllers[3].GetChannels()

	roomChannels[1].SetConditions([]config.Condition{
		config.Condition{
			MetricID:   roomController.GetMetrics()[1].GetID(),
			Comparator: ">",
			Threshold:  85.0}})

	roomChannels[2].SetConditions([]config.Condition{
		config.Condition{
			MetricID:   roomController.GetMetrics()[1].GetID(),
			Comparator: "<",
			Threshold:  70.0}})

	roomChannels[3].SetConditions([]config.Condition{
		config.Condition{
			MetricID:   roomController.GetMetrics()[3].GetID(),
			Comparator: ">",
			Threshold:  55.0}})

	roomChannels[5].SetConditions([]config.Condition{
		config.Condition{
			MetricID:   roomController.GetMetrics()[14].GetID(),
			Comparator: "<",
			Threshold:  1200.0}})

	reservoirChannels[1].SetConditions([]config.Condition{
		config.Condition{
			MetricID:   reservoirController.GetMetrics()[2].GetID(),
			Comparator: ">",
			Threshold:  62.0}})

	reservoirChannels[2].SetConditions([]config.Condition{
		config.Condition{
			MetricID:   reservoirController.GetMetrics()[2].GetID(),
			Comparator: "<",
			Threshold:  60.0}})

	doserChannels[0].SetConditions([]config.Condition{
		config.Condition{
			MetricID:   reservoirController.GetMetrics()[2].GetID(),
			Comparator: ">",
			Threshold:  6.1}})

	doserChannels[1].SetConditions([]config.Condition{
		config.Condition{
			MetricID:   reservoirController.GetMetrics()[2].GetID(),
			Comparator: "<",
			Threshold:  5.4}})

	doserChannels[2].SetConditions([]config.Condition{
		config.Condition{
			MetricID:   reservoirController.GetMetrics()[5].GetID(),
			Comparator: "<",
			Threshold:  300.0}})

	farmDAO.Save(farmConfig.(*config.Farm))

	initializer.seedInventory(db)

	return nil
}

func (initializer *ClusterInitializer) seedInventory(db *gorm.DB) {
	db.Create(&entity.InventoryType{
		Name:             "pH Probe",
		Description:      "+/- 0.1 resolution",
		Image:            "https://www.atlas-scientific.com/_images/large_images/probes/c-ph-box.png",
		LifeExpectancy:   47336400, // 1.5 yrs
		MaintenanceCycle: 31557600, // 1 yr
		ProductPage:      "https://www.atlas-scientific.com/product_pages/probes/c-ph-probe.html"})
	db.Create(&entity.InventoryType{
		Name:             "Oxygen Reduction Potential Probe (ORP)",
		Description:      "+/– 1.1mV accuracy",
		Image:            "https://www.atlas-scientific.com/_images/large_images/probes/c-orp-box.png",
		MaintenanceCycle: 31557600, // 1 yr
		LifeExpectancy:   47336400, // 1.5 yrs
		ProductPage:      "https://www.atlas-scientific.com/product_pages/probes/c-orp-probe.html"})
	db.Create(&entity.InventoryType{
		Name:             "Conductivity Probe (EC/TDS)",
		Description:      "+/ – 2% accuracy",
		Image:            "https://www.atlas-scientific.com/_images/large_images/probes/ec_k0-1.png",
		LifeExpectancy:   315576000, // 10 yrs
		MaintenanceCycle: 315576000,
		ProductPage:      "https://www.atlas-scientific.com/product_pages/probes/ec_k0-1.html"})
	db.Create(&entity.InventoryType{
		Name:             "Dissolved Oxygen Probe (DO)",
		Description:      "+/ – 2% accuracy",
		Image:            "https://www.atlas-scientific.com/_images/large_images/probes/do_box.png",
		LifeExpectancy:   157788000, // 5 yrs
		MaintenanceCycle: 47336400,  // 18 months
		ProductPage:      "https://www.atlas-scientific.com/product_pages/probes/do_probe.html"})
	db.Create(&entity.InventoryType{
		Name:           "NDIR Carbon Dioxide Sensor (CO2)",
		Description:    "+/- 3%  +/- 30 ppm",
		Image:          "https://dqzrr9k4bjpzk.cloudfront.net/images/805759/897784873.jpg",
		LifeExpectancy: 173566800, // 5.5 yrs
		ProductPage:    "https://www.atlas-scientific.com/product_pages/probes/ezo-co2.html"})
	db.Create(&entity.InventoryType{
		Name:           "pH Calibration Fluids",
		Description:    "3x NIST reference calibrated 500ml bottles (4.0, 7.0, 10.0)",
		Image:          "https://images-na.ssl-images-amazon.com/images/I/81HsYv4kNTL._SL1500_.jpg",
		LifeExpectancy: 31557600, // 1 yr
		ProductPage:    "https://www.amazon.com/Biopharm-Calibration-Solution-Standards-Traceable/dp/B01E7U873K"})
	db.Create(&entity.InventoryType{
		Name:           "pH UP (BASE)",
		Description:    "General Hydroponics pH Up Liquid Fertilizer, 1-Gallon",
		Image:          "https://images-na.ssl-images-amazon.com/images/I/61mHEr-obpL._AC_SL1200_.jpg",
		LifeExpectancy: 31557600, // 1 yr
		ProductPage:    "https://www.amazon.com/General-Hydroponics-Liquid-Fertilizer-1-Gallon/dp/B000FF5V90"})
	db.Create(&entity.InventoryType{
		Name:           "pH DOWN (ACID)",
		Description:    "General Hydroponics HGC722125 Liquid Premium Buffering for pH Stability, 1-Gallon, Orange",
		Image:          "https://images-na.ssl-images-amazon.com/images/I/71E-fJ-tlsL._AC_SL1500_.jpg",
		LifeExpectancy: 31557600, // 1 yr
		ProductPage:    "https://www.amazon.com/General-Hydroponics-Liquid-Fertilizer-1-Gallon/dp/B000FG0F9U"})
}
