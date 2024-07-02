package packet

import "gladiatorsGoModule/setting"

// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"

type BattleState struct {
	CMDContent
}

type BattleState_ToClient struct {
	CMDContent
	PlayerStates [][setting.PLAYER_NUMBER]PackPlayerState
	GameTime     []float64
}

type PackPlayerState struct {
	ID          string // 玩家DBID
	BribeSkills [2]PackBribeSkill
	Gladiator   PackGladiator
}

type PackBribeSkill struct {
	JsonID int
	Used   bool
}

type PackGladiator struct {
	LeftSide        bool
	JsonGladiatorID int
	JsonSkillIDs    [6]int
	JsonTraitIDs    []int
	JsonEquipIDs    []int
	CurSkills       [4]PackSkill
	HP              int
	CurHP           int
	CurVigor        float64
	VigorRegen      float64
	STR             int
	DEF             int
	MDEF            int
	CRIT            float64
	INIT            int
	Knockback       int
	Speed           int     // 當前速率(N表示每秒前進N*GridUnit個Unit)
	BattlePos       int     // 戰鬥位置(Server計算用)
	StagePos        float64 // 場景上實際位置(Client表演用)
	Buffers         []PackBuffer
}

type PackSkill struct {
	JsonID int
	On     bool
}

type PackBuffer struct {
	JsonID string
	Stack  int
}
