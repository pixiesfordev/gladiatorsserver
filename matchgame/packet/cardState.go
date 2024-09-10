package packet

import (
// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"
)

type CardState_ToClient struct {
	CMDContent
	MyCardState PackCardState
}

type PackCardState struct {
	HandSkillIDs    [4]int // 手牌技能
	HandOnID        int    // 啟用的手牌技能
	DivineSkillIDs  [2]int // 神祉技能
	DivineSkillOnID int    // 啟用的神祉技能ID
}
