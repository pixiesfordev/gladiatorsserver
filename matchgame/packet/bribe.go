package packet

import "gladiatorsGoModule/setting"

// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"

type Bribe struct {
	CMDContent
	JsonBribeIDs [2]int
}

type Bribe_ToClient struct {
	CMDContent
	Players  [setting.PLAYER_NUMBER]PackPlayerState
	GameTime float64
}
