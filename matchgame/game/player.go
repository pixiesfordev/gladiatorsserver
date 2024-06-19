package game

import (
	// "fmt"
	// "gladiatorsGoModule/gameJson"
	// "gladiatorsGoModule/utility"
	"gladiatorsGoModule/setting"
	"matchgame/logger"
	"matchgame/packet"
	"time"

	log "github.com/sirupsen/logrus"
)

type Gamer interface {
	SetIdx(idx int)
	GetID() string
	GetGold() int64
	AddGold(value int64)
	GetGladiator() *Gladiator
	IsReady() bool
	GetPackPlayerState() packet.PackPlayerState
	GetPackPlayerBribes() [setting.PLAYER_NUMBER]packet.PackBribeSkill
}

// 玩家
type Player struct {
	ID           string                         // DBPlayer的_id
	Idx          int                            // 第一位玩家是0(左方) 第二位玩家是1(右方)
	myGladiator  *Gladiator                     // 使用中的鬥士
	gold         int64                          // 玩家金幣
	ready        bool                           // 是否準備好了(進遊戲且收到雙方玩家資料後, client會送準備封包設定ready為true)
	BribeSkills  [DivineSkillCount]*DivineSkill // 神祉技能
	LastUpdateAt time.Time                      // 上次收到玩家更新封包(心跳)
	ConnTCP      *ConnectionTCP                 // TCP連線
	ConnUDP      *ConnectionUDP                 // UDP連線
}

func (player *Player) SetIdx(idx int) {
	player.Idx = idx
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

func (player *Player) GetPackPlayerBribes() [setting.PLAYER_NUMBER]packet.PackBribeSkill {
	var playerBribes [2]packet.PackBribeSkill

	playerBribes[0] = packet.PackBribeSkill{
		JsonID: player.BribeSkills[0].MyJson.ID,
		Used:   player.BribeSkills[0].Used,
	}
	playerBribes[1] = packet.PackBribeSkill{
		JsonID: player.BribeSkills[1].MyJson.ID,
		Used:   player.BribeSkills[1].Used,
	}

	return playerBribes
}

func (player *Player) GetPackPlayerState() packet.PackPlayerState {
	packPlayerState := packet.PackPlayerState{
		ID:          player.GetID(),
		BribeSkills: player.GetPackPlayerBribes(),
		Gladiator:   player.GetGladiator().GetPackGladiator(),
	}
	return packPlayerState
}
