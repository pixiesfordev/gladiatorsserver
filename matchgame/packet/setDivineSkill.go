package packet

// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"

type SetDivineSkill struct {
	CMDContent
	JsonSkillIDs [2]int
}

type SetDivineSkill_ToClient struct {
	CMDContent
	MyPlayerState       PackPlayerState
	MyCardState         PackCardState
	OpponentPlayerState PackPlayerState
}
