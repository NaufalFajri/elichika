package client

import (
	"elichika/generic"
)

type UserInfoTriggerBasic struct {
	TriggerId       int64                    `xorm:"pk 'trigger_id'" json:"trigger_id"` // use nano timestamp
	InfoTriggerType int32                    `json:"info_trigger_type" enum:"InfoTriggerType"`
	LimitAt         generic.Nullable[int64]  `xorm:"json 'limit_at'" json:"limit_at"`       // seems like some sort of timed timestamp, probably for event popup
	Description     generic.Nullable[string] `xorm:"json 'description'" json:"description"` // this is a string that can be null
	ParamInt        generic.Nullable[int32]  `xorm:"json 'param_int'" json:"param_int"`
}
