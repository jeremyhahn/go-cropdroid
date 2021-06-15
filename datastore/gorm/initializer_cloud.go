package gorm

import (
	"time"

	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/gorm/entity"
	"github.com/jinzhu/gorm"
	logging "github.com/op/go-logging"
)

type ClusterInitializerGossipTest struct {
	logger   *logging.Logger
	db       *gorm.DB
	location *time.Location
	DatabaseInitializer
}

func NewGormCloudInitializer(logger *logging.Logger, db *gorm.DB, location *time.Location) DatabaseInitializer {
	return &ClusterInitializerGossipTest{
		logger:   logger,
		db:       db,
		location: location}
}

func (initializer *ClusterInitializerGossipTest) Initialize() error {

	db := initializer.db

	db.LogMode(true)

	db.AutoMigrate(&config.Permission{})
	db.AutoMigrate(&config.User{})
	db.AutoMigrate(&config.Role{})
	db.AutoMigrate(&config.Device{})
	db.AutoMigrate(&config.DeviceConfigItem{})
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

	return nil
}
