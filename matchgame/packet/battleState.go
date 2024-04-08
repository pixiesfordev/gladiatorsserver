package packet

import (
// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"
)

type BattleState struct {
	CMDContent
}

type BattleState_ToClient struct {
	CMDContent
	Players  [2]PackPlayerState
	GameTime float64
}

type PackPlayerState struct {
	ID          string // 玩家DBID
	BribeSkills [2]PackBribeSkill
	Gladiator  PackGladiator
}

type PackBribeSkill struct {
	JsonID int
	Used   bool
}

type PackGladiator struct {
	JsonGladiatorID int
	JsonSkillIDs    [6]int
	JsonTraitIDs    []int
	JsonEquipIDs    []int
	CurJsonSkillIDs [4]PackSkill
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
	BattlePos       int        // 戰鬥位置
	StagePos        [2]float64 // 場景上實際位置
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
