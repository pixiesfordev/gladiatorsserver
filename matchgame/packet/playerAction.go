package packet

// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"

type ActionType string

const (
	ACTION_SKILL         ActionType = "ACTION_SKILL"         // 啟用技能
	ACTION_OPPONENTSKILL            = "ACTION_OPPONENTSKILL" // 衝刺
	ACTION_RUSH                     = "ACTION_RUSH"          // 衝刺
	ACTION_DIVINESKILL              = "ACTION_DIVINESKILL"   // 啟用神祉技能
	ACTION_SURRENDER                = "ACTION_SURRENDER"     // 投降
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
type PackAction_Skill_ToClient struct {
	SkillOnID    int
	HandSkillIDs [4]int
}
type PackAction_OpponentSkill_ToClient struct {
	SkillID int
}

// 施放神祉技能
type PackAction_DivineSkill struct {
	On      bool
	SkillID int
}

type PackAction_DivineSkill_ToClient struct {
	On      bool
	SkillID int
}

// 衝刺
type PackAction_Rush struct {
	On bool
}

// 投降
type PackAction_Surrender struct {
}
type PackAction_Surrender_ToClient struct {
}
