package util

import "github.com/jeremyhahn/go-cropdroid/common"

type SwitchPosition struct {
	position int
}

func NewSwitchPosition(position int) *SwitchPosition {
	return &SwitchPosition{position: position}
}

func (sp *SwitchPosition) ToString() string {
	if sp.position == common.SWITCH_ON {
		return "ON"
	} else {
		return "OFF"
	}
}
