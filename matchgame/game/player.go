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

func (player *Player) SetReady() {
	player.ready = true
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

func (player *Player) GetPackDivineSkills() [setting.PLAYER_NUMBER]packet.PackDivineSkill {
	var playerBribes [2]packet.PackDivineSkill
	return playerBribes
}

// GetPackPlayerState 取得玩家的狀態封包
func (player *Player) GetPackPlayer() packet.PackPlayer {
	packPlayer := packet.PackPlayer{
		DBID: player.GetID(),
	}
	return packPlayer
}

// GetPackPlayerState 取得玩家的狀態封包
func (player *Player) GetOpponentPackPlayer() packet.PackPlayer {
	if LeftGamer.GetID() == player.ID {
		return (RightGamer.(*Player)).GetPackPlayer()
	} else {
		return (LeftGamer.(*Player)).GetPackPlayer()
	}
}

// GetPackPlayerState 取得玩家的狀態封包
func (player *Player) GetPackPlayerState() packet.PackPlayerState {
	packPlayerState := packet.PackPlayerState{
		DBID:           player.GetID(),
		DivineSkills:   player.GetPackDivineSkills(),
		GladiatorState: player.GetGladiator().GetPackGladiatorState(),
	}
	return packPlayerState
}

// GetOpponentPackPlayerState 取得玩家對手的狀態封包
func (player *Player) GetOpponentPackPlayerState() packet.PackPlayerState {
	if LeftGamer.GetID() == player.ID {
		return (RightGamer.(*Player)).GetPackPlayerState()
	} else {
		return (LeftGamer.(*Player)).GetPackPlayerState()
	}
}
