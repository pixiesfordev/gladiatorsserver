package packet

// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"

type ActionType string

const (
	Action_Skill       ActionType = "Action_Skill"       // 啟用技能
	Action_Rush                   = "Action_Rush"        // 衝刺
	Action_DivineSkill            = "Action_DivineSkill" // 啟用神祉技能
)

// 玩家動作
type PlayerAction struct {
	CMDContent
	ActionType    ActionType
	ActionContent interface{}
}
type PlayerAction_ToClient struct {
	CMDContent
	PlayerDBID    string
	ActionType    ActionType
	ActionContent interface{}
}

// 施放技能
type PackAction_Skill struct {
	On      bool
	SkillID int
}

// 施放神祉技能
type PackAction_BribeSkill struct {
	On      bool
	SkillID int
}

// 衝刺
type PackAction_Rush struct {
	On bool
}

// 衝刺ToClient
type PackAction_Rush_ToClient struct {
	On bool
}

// 投降
type PackAction_Surrender struct {
}
