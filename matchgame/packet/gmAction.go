package packet

// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"

type GMActionType string

const (
	GMACTION_SETSKILLS GMActionType = "GMACTION_SETSKILLS"
)

type GMAction struct {
	CMDContent
	ActionType    GMActionType
	ActionContent interface{}
}
type GMAction_ToClient struct {
	CMDContent
	PlayerDBID    string
	ActionType    GMActionType
	Result        bool
	ActionContent interface{}
}

// 施放技能
type PackGMAction_SetSkills struct {
	SkillIDs [6]int
}
