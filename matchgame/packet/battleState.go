package packet

import (
	// logger "matchgame/logger"
	// log "github.com/sirupsen/logrus"
	"gladiatorsGoModule/utility"
)

type BattleState struct {
	CMDContent
}

type BattleState_ToClient struct {
	CMDContent
	Players  []PackPlayerState
	GameTime float64
}

type PackPlayerState struct {
	Skills      []PackSkill
	BribeSkills []PackBribeSkill
	Gladiators  []PackGladiator
}

type PackSkill struct {
	JsonID int
	On     bool
}

type PackBribeSkill struct {
	JsonID int
	Used   bool
}

type PackGladiator struct {
	HP        int
	Vigor     int
	BattlePos int             // 戰鬥位置
	StagePos  utility.Vector2 // 場景上實際位置
	Buffers   []PackBuffer
}
type PackBuffer struct {
	JsonID string
	Stack  int
}
