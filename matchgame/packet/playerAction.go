package packet

import "gladiatorsGoModule/setting"

// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"

// 玩家動作
type PlayerAction struct {
	CMDContent
	ActionType    string
	ActionContent interface{}
}
type PlayerAction_ToClient struct {
	CMDContent
	ActionType    string
	ActionContent interface{}
	PlayerStates  [][setting.PLAYER_NUMBER]PackPlayerState
	GameTime      []float64
}

// 施放技能
type PackAction_Skill struct {
	On       bool
	SkillIdx int
}

// 施放神祉技能
type PackAction_BribeSkill struct {
	On            bool
	BribeSkillIdx int
}

// 衝刺
type PackAction_Rush struct {
	On bool
}

// 投降
type PackAction_Surrender struct {
}
