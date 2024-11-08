package packet

// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"

type GMActionType string

const (
	GMACTION_SETGLADIATOR GMActionType = "GMACTION_SETGLADIATOR"
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
	ActionContent interface{}
}

// 設定角鬥士
type PackGMAction_SetGladiator struct {
	GladiatorID int
	SkillIDs    [6]int
}
