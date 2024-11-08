package packet

// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"

type ActionType string

const (
	ACTION_SKILL       ActionType = "ACTION_SKILL"       // 啟用技能
	ACTIVE_MELEE_SKILL ActionType = "ACTIVE_MELEE_SKILL" // (ACTION_SKILL的回傳Client)
	INSTANT_SKILL      ActionType = "INSTANT_SKILL"      // (ACTION_SKILL的回傳Client)
	ACTION_RUSH        ActionType = "ACTION_RUSH"        // 衝刺
	ACTION_DIVINESKILL ActionType = "ACTION_DIVINESKILL" // 啟用神祉技能
	ACTION_SURRENDER   ActionType = "ACTION_SURRENDER"   // 投降
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

// 啟用肉搏技能
type PackAction_ActiveMeleeSkill_ToClient struct {
	On      bool
	SkillID int
}

// 發動立即技能
type PackAction_InstantSkill_ToClient struct {
	SkillID    int    // 施放的即時技能ID
	NewSkilID  int    // 新抽到的技能(對手不會收到)
	HandSkills [4]int // 手牌(對手不會收到)
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
