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
	IsSelectedDivineSkill() bool
	GetPackPlayer(myself bool) packet.PackPlayer
	GetPackPlayerState(myselfPack bool) packet.PackPlayerState
}

// 玩家
type Player struct {
	ID                  string                         // DBPlayer的_id
	Idx                 int                            // 第一位玩家是0(左方) 第二位玩家是1(右方)
	MyGladiator         *Gladiator                     // 使用中的鬥士
	gold                int64                          // 玩家金幣
	ready               bool                           // 是否準備好了(進遊戲且收到雙方玩家資料後, client會送準備封包設定ready為true)
	selectedDivineSkill bool                           // 是否已經選好神祉技能(雙方都選好神祉技能, 或倒數結束就會進入戰鬥)
	DivineSkills        [DivineSkillCount]*DivineSkill // 神祉技能
	LastUpdateAt        time.Time                      // 上次收到玩家更新封包(心跳)
	ConnTCP             *ConnectionTCP                 // TCP連線
	ConnUDP             *ConnectionUDP                 // UDP連線
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
	return player.MyGladiator
}

func (player *Player) SetReady() {
	player.ready = true
}

func (player *Player) IsReady() bool {
	return player.ready
}

func (player *Player) SetSelectedDivineSkill() {
	player.selectedDivineSkill = false
}

func (player *Player) IsSelectedDivineSkill() bool {
	return player.selectedDivineSkill
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
	var packDivineSkills [2]packet.PackDivineSkill
	return packDivineSkills
}

// GetPackPlayer 取得玩家封包
func (player *Player) GetPackPlayer(myself bool) packet.PackPlayer {
	packPlayer := packet.PackPlayer{
		DBID:            player.GetID(),
		MyPackGladiator: player.MyGladiator.GetPackGladiator(myself),
	}
	return packPlayer
}

// GetOpponent 取得對手Gamer
func (player *Player) GetOpponent() Gamer {
	if LeftGamer.GetID() == player.ID {
		return RightGamer
	} else {
		return LeftGamer
	}
}

// GetPackPlayerState 取得玩家的狀態封包
func (player *Player) GetPackPlayerState(myselfPack bool) packet.PackPlayerState {
	packPlayerState := packet.PackPlayerState{
		DBID:           player.GetID(),
		DivineSkills:   player.GetPackDivineSkills(),
		GladiatorState: player.GetGladiator().GetPackGladiatorState(myselfPack),
	}
	return packPlayerState
}

// GetOpponentPackPlayerState 取得玩家對手的狀態封包
func (player *Player) GetOpponentPackPlayerState() packet.PackPlayerState {
	if LeftGamer.GetID() == player.ID {
		return RightGamer.GetPackPlayerState(false)
	} else {
		return LeftGamer.GetPackPlayerState(false)
	}
}

// 送封包給玩家(TCP)
func (p *Player) SendPacketToPlayer(pack packet.Pack) {
	if p.ConnTCP.Conn == nil {
		return
	}
	err := packet.SendPack(p.ConnTCP.Encoder, pack)
	if err != nil {
		log.Errorf("%s SendPacketToPlayer error: %v", logger.LOG_Room, err)
	}
}
