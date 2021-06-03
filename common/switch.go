package common

type Switch struct {
	Channel int `json:"channel"`
	Pin     int `json:"pin"`
	State   int `json:"position"`
}

func (_switch *Switch) GetChannel() int {
	return _switch.Channel
}

func (_switch *Switch) GetPin() int {
	return _switch.Pin
}

func (_switch *Switch) GetState() int {
	return _switch.State
}
