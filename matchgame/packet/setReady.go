package packet

// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"

type SetReady struct {
	CMDContent
}

type SetReady_ToClient struct {
	CMDContent
	PlayerReadies [2]bool
}
