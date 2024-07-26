package packet

// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"

type DivineSkill struct {
	CMDContent
	JsonSkillIDs [2]int
}

type DivineSkill_ToClient struct {
	CMDContent
	MyPlayerState       PackPlayerState
	OpponentPlayerState PackPlayerState
}
