package game

import (
	// "fmt"
	// "gladiatorsGoModule/gameJson"
	// "gladiatorsGoModule/utility"
	"matchgame/logger"
	gSetting "matchgame/setting"
	"time"

	log "github.com/sirupsen/logrus"
)

type Gamer interface {
	GetID() string
	GetGold() int64
	AddGold(value int64)
	GetGladiator() *Gladiator
	IsReady() bool
}

// 玩家
type Player struct {
	ID           string                  // DBPlayer的_id
	myGladiator  *Gladiator              // 使用中的鬥士
	gold         int64                   // 玩家金幣
	ready        bool                    // 是否準備好了(進遊戲且收到雙方玩家資料後, client會送準備封包設定ready為true)
	LastUpdateAt time.Time               // 上次收到玩家更新封包(心跳)
	ConnTCP      *gSetting.ConnectionTCP // TCP連線
	ConnUDP      *gSetting.ConnectionUDP // UDP連線
}

func (player *Player) GetID() string {
	return player.ID
}

func (player *Player) GetGold() int64 {
	return player.gold
}

func (player *Player) AddGold(value int64) {
	player.gold += value
}

func (player *Player) GetGladiator() *Gladiator {
	return player.myGladiator
}

func (player *Player) IsReady() bool {
	return player.ready
}

// 將玩家連線斷掉
func (player *Player) CloseConnection() {
	if player == nil {
		log.Errorf("%s 關閉玩家連線時 player 為 nil", logger.LOG_Player)
		return
	}
	if player.ConnTCP.Conn != nil {
		player.ConnTCP.MyLoopChan.ClosePackReadStopChan()
		player.ConnTCP.Conn.Close()
		player.ConnTCP.Conn = nil
		player.ConnTCP = nil
	}
	if player.ConnUDP.Conn != nil {
		player.ConnUDP.MyLoopChan.ClosePackReadStopChan()
		player.ConnUDP.Conn = nil
		player.ConnUDP = nil
	}
	log.Infof("%s 關閉玩家(%s)連線", logger.LOG_Player, player.GetID())
}
