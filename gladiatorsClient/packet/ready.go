package packet

// logger "gladiatorsClient/logger"
// log "github.com/sirupsen/logrus"

type Ready struct {
	CMDContent
}

type Ready_ToClient struct {
	CMDContent
	PlayerReadies [2]bool
}
