package packet

import (
// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"
)

type Bribe struct {
	CMDContent
	JsonBribeIDs [2]int
}

type Bribe_ToClient struct {
	CMDContent
	PlayerBribes [2]PackPlayerBribe
}
type PackPlayerBribe struct {
	JsonBribeIDs [2]int
}
