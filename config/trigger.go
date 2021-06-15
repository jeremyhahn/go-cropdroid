package config

type Trigger struct {
	Device  string `yaml:"device" json:"device"`
	Channel int    `yaml:"channel" json:"channel"`
	State   int    `yaml:"state" json:"state"`
	Timer   string `yaml:"timer" json:"timer"`
	Async   bool   `yaml:"async" json:"async"`
	Wait    bool   `yaml:"wait" json:"wait"`
}
