package game

import (
	// "fmt"
	// "gladiatorsGoModule/gameJson"
	// "gladiatorsGoModule/utility"
	// "gladiatorsGoModule/setting`"
	"matchgame/logger"
	"matchgame/packet"
	"sync"
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
	Surrender()
	GetPackPlayer(myself bool) packet.PackPlayer
	GetOpponent() Gamer
	SetOpponent(gamer Gamer)
}

// 玩家
type Player struct {
	ID                        string                         // DBPlayer的_id
	Idx                       int                            // 第一位玩家是0(左方) 第二位玩家是1(右方)
	opponent                  Gamer                          // 對手
	MyGladiator               *Gladiator                     // 使用中的鬥士
	gold                      int64                          // 玩家金幣
	ready                     bool                           // 是否準備好了(進遊戲且收到雙方玩家資料後, client會送準備封包設定ready為true)
	selectDivineSkillFinished bool                           // 是否已經選好神祉技能(雙方都選好神祉技能, 或倒數結束就會進入戰鬥)
	DivineSkills              [DivineSkillCount]*DivineSkill // 神祉技能
	LastUpdateAt              time.Time                      // 上次收到玩家更新封包(心跳)
	ConnTCP                   *ConnectionTCP                 // TCP連線
	ConnUDP                   *ConnectionUDP                 // UDP連線
	MutexLock                 sync.Mutex
}

func NewPlayer(id string, connTCP *ConnectionTCP, connUDP *ConnectionUDP) *Player {
	player := &Player{
		ID:           id,
		LastUpdateAt: time.Now(),
		ConnTCP:      connTCP,
		ConnUDP:      connUDP,
	}
	return player
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
func (player *Player) Surrender() {

}

func (player *Player) IsReady() bool {
	return player.ready
}

func (player *Player) FinishSelectedDivineSkill() {
	player.selectDivineSkillFinished = false
}

func (player *Player) IsSelectedDivineSkill() bool {
	return player.selectDivineSkillFinished
}

// 將玩家連線斷掉
func (player *Player) CloseConnection() {
	if player == nil {
		log.Errorf("%s 關閉玩家連線時 player 為 nil", logger.LOG_Player)
		return
	}
	player.MutexLock.Lock()
	defer player.MutexLock.Unlock()

	if player.ConnTCP.Conn != nil {
		player.ConnTCP.MyLoopChan.Close()
		player.ConnTCP.Conn.Close()
		player.ConnTCP.Conn = nil
		player.ConnTCP = nil
	}
	if player.ConnUDP.Conn != nil {
		player.ConnUDP.MyLoopChan.Close()
		player.ConnUDP.Conn = nil
		player.ConnUDP = nil
	}
	log.Infof("%s 關閉玩家(%s)連線", logger.LOG_Player, player.GetID())
}

// GetOpponent 取得對手Gamer
func (player *Player) GetOpponent() Gamer {
	return player.opponent
}

// SetOpponent 設定對手Gamer
func (player *Player) SetOpponent(gamer Gamer) {
	player.opponent = gamer
}

// GetPackPlayer 取得玩家封包
func (player *Player) GetPackPlayer(myself bool) packet.PackPlayer {
	packPlayer := packet.PackPlayer{
		DBID:            player.GetID(),
		MyPackGladiator: player.MyGladiator.GetPackGladiator(myself),
	}
	return packPlayer
}

// GetOpponentPackPlayer 取得對手封包
func (player *Player) GetOpponentPackPlayer(myself bool) packet.PackPlayer {
	opponent := player.GetOpponent()
	if opponent != nil {
		return opponent.GetPackPlayer(myself)
	}
	return packet.PackPlayer{}
}

// 送封包給玩家(TCP)
func (p *Player) SendPacketToPlayer(pack packet.Pack) {
	if p.ConnTCP.Conn == nil {
		return
	}
	p.MutexLock.Lock()
	defer p.MutexLock.Unlock()
	err := packet.SendPack(p.ConnTCP.Encoder, pack)
	if err != nil {
		log.Errorf("%s SendPacketToPlayer error: %v", logger.LOG_Room, err)
	}
}
