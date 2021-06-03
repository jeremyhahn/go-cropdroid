package yaml

type Server struct {
	Interval      int            `yaml:"interval" json:"interval"`
	Timezone      string         `yaml:"timezone" json:"timezone"`
	Mode          string         `yaml:"mode" json:"mode"`
	Smtp          Smtp           `yaml:"smtp" json:"smtp"`
	License       License        `yaml:"license" json:"license"`
	Organizations []Organization `yaml:"organizations" json:"organizations"`
}

type Smtp struct {
	Enable    bool   `yaml:"enable" json:"enable"`
	Host      string `yaml:"host" json:"host"`
	Port      int    `yaml:"port" json:"port"`
	Username  string `yaml:"username" json:"username"`
	Password  string `yaml:"password" json:"password"`
	Recipient string `yaml:"recipient" json:"recipient"`
}

type Organization struct {
	ID      int     `gorm:"primary_key;AUTO_INCREMENT" yaml:"id" json:"id"`
	Name    string  `gorm:"size:255" yaml:"name" json:"name"`
	Farms   []Farm  `yaml:"farms" json:"farms"`
	Users   []User  `yaml:"users" json:"users"`
	License License `yaml:"license" json:"license"`
}

type Farm struct {
	ID       int    `yaml:"id" json:"id"`
	OrgID    int    `yaml:"orgId" json:"orgId"`
	Mode     string `yaml:"mode" json:"mode"`
	Name     string `yaml:"name" json:"name"`
	Interval int    `yaml:"interval" json:"interval"`
	Users    []User `yaml:"users" json:"users"`
}

type User struct {
	ID       int      `yaml:"id" json:"id"`
	Email    string   `yaml:"email" json:"email"`
	Password string   `yaml:"password" json:"password"`
	Roles    []string `yaml:"roles" json:"roles"`
}

type License struct {
	UserQuota       int `yaml:"userQuota" json:"userQuota"`
	FarmQuota       int `yaml:"farmQuota" json:"farmQuota"`
	ControllerQuota int `yaml:"controllerQuota" json:"controllerQuota"`
}
