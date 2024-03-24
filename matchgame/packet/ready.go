package packet

import (
// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"
)

type Ready struct {
	CMDContent
}

type Ready_ToClient struct {
	CMDContent
	PlayerReadys [2]bool
}
