package common

import "time"

const (
	APPNAME               = "cropdroid"
	DISPLAYNAME           = "CropDroid™"
	PACKAGE               = "com.jeremyhahn.cropdroid"
	TIME_FORMAT           = time.RFC3339
	TIME_FORMAT_LOCAL     = "2006-01-02T15:04:05Z"
	TIME_DISPLAY_FORMAT   = "01-02-2006 15:04:05 MST"
	TIME_RFC1123_FORMAT   = "Mon, 02 Jan 2006 15:04:05 MST"
	TIME_NULL             = "0001-01-01 00:00:00+00:00"
	BUFFERED_CHANNEL_SIZE = 256
	WEBSOCKET_KEEPALIVE   = 10 * time.Second
	HTTP_CLIENT_TIMEOUT   = 10 * time.Second
	HTTP_PUBLIC_HTML      = "public_html"
	SWITCH_OFF            = 0
	SWITCH_ON             = 1
	//FARM_MAX_SIZE         = 10000
	HOURS_IN_A_YEAR                    = 8766
	DEFAULT_FARM_CONFIG_HISTORY_LENGTH = 5

	DATATYPE_FLOAT  = 0
	DATATYPE_INT    = 1
	DATATYPE_STRING = 2

	NOTIFICATION_PRIORITY_LOW  = 0
	NOTIFICATION_PRIORITY_MED  = 1
	NOTIFICATION_PRIORITY_HIGH = 2

	// IDs of devices as inserted by dao.Initializer
	CONTROLLER_TYPE_ID_SERVER    = 1
	CONTROLLER_TYPE_ID_ROOM      = 2
	CONTROLLER_TYPE_ID_RESERVOIR = 3
	CONTROLLER_TYPE_ID_DOSER     = 4

	CONTROLLER_TYPE_ROOM      = "room"
	CONTROLLER_TYPE_DOSER     = "doser"
	CONTROLLER_TYPE_RESERVOIR = "reservoir"
	CONTROLLER_TYPE_SERVER    = "server"

	CONFIG_NAME_KEY     = "name"
	CONFIG_INTERVAL_KEY = "interval"
	CONFIG_TIMEZONE_KEY = "timezone"
	CONFIG_MODE_KEY     = "mode"

	CONFIG_SMTP_ENABLE_KEY    = "smtp.enable"
	CONFIG_SMTP_HOST_KEY      = "smtp.host"
	CONFIG_SMTP_PORT_KEY      = "smtp.port"
	CONFIG_SMTP_USERNAME_KEY  = "smtp.username"
	CONFIG_SMTP_PASSWORD_KEY  = "smtp.password"
	CONFIG_SMTP_RECIPIENT_KEY = "smtp.recipient"

	CONFIG_MODE_VIRTUAL     = "virtual"
	CONFIG_MODE_SERVER      = "server"
	CONFIG_MODE_CLOUD       = "cloud"
	CONFIG_MODE_MAINTENANCE = "maintenance"

	SCHEDULE_FREQUENCY_DOES_NOT_REPEAT = 0
	SCHEDULE_FREQUENCY_DAILY           = 1
	SCHEDULE_FREQUENCY_WEEKLY          = 2
	SCHEDULE_FREQUENCY_MONTHLY         = 3
	SCHEDULE_FREQUENCY_YEARLY          = 4

	ALGORITHM_PH_ID  = 1
	ALGORITHM_ORP_ID = 2

	MODE_STANDALONE = "standalone"
	MODE_CLUSTER    = "cluster"
	MODE_CLOUD      = "cloud"

	DEFAULT_USER     = "admin"
	DEFAULT_PASSWORD = "cropdroid"
	DEFAULT_ROLE     = "admin"

	AUTH_TYPE_LOCAL  = 0
	AUTH_TYPE_GOOGLE = 1
)
