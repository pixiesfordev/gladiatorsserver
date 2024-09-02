package game

import (
	"fmt"
	"gladiatorsGoModule/gameJson"
	"gladiatorsGoModule/utility"
	"math"
	"net"

	"encoding/json"
	logger "matchgame/logger"
	"matchgame/packet"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type GameState string // 目前遊戲狀態列舉
const (
	GameState_Initializing         GameState = "GameState_Initializing"
	GameState_Inited                         = "GameState_Inited"
	GameState_WaitingPlayers                 = "GameState_WaitingPlayers"       // 等待雙方玩家入場
	GameState_SelectingDivineSkill           = "GameState_SelectingDivineSkill" // 選擇神祉技能
	GameState_CountingDown                   = "GameState_CountingDown"         // 戰鬥倒數開始中
	GameState_Fighting                       = "GameState_Fighting"             // 戰鬥中
	GameState_End                            = "GameState_End"                  // 結束戰鬥
)

const (
	PingMiliSecs                           = 1000 // 每X毫秒送Ping封包給client(心跳檢測)
	AGONES_HEALTH_PIN_INTERVAL_SEC         = 2    // 每X秒檢查AgonesServer是否正常運作(官方文件範例是用2秒)
	TCP_CONN_TIMEOUT_SEC                   = 120  // TCP連線逾時時間X秒
	BattleLOOP_MILISECS            int     = 100  // 戰鬥每X毫秒循環
	GameLOOP_MILISECS              int     = 1000 // 遊戲每X毫秒循環
	KICK_PLAYER_SECS               float64 = 10   // 最長允許玩家無心跳X秒後踢出遊戲房
	MarketDivineSkillCount                 = 4    // 有幾個神祉技能可以購買
	DivineSkillCount                       = 2    // 玩家可以買幾個神祉技能
	GladiatorSkillCount                    = 6    // 玩家有幾個技能
	HandSkillCount                         = 4    // 玩家手牌技能, 索引0的技能是下一張牌
	WAIT_BATTLE_START                      = 2    // (測試用)BattleStart等待時間
	CollisionDis                           = 4    // 相距X單位就算碰撞
	MaxVigor                       float64 = 20   // 最大體力
	SelectDivineCountDownSecs      int     = 15   // 選神祉技能倒數秒數
	FightingCountDownSecs          int     = 4    // 戰鬥倒數秒數
)

// 戰鬥
const (
	WallPos          = 20.0 // 中心點距離牆壁的單位數, 填20代表中心點左右20單位是牆壁, 也就是戰鬥場地總長度是40
	InitGladiatorPos = 16.0 // 雙方角鬥士初始位置, 填16代表, 角鬥士戰鬥開始的起始位置距離中心點16單位
)

var IDAccumulator = utility.NewAccumulator() // 產生一個ID累加器
// Mode模式分為以下:
// standard:一般版本
// non-agones: 個人測試模式(不使用Agones服務, non-agones的連線方式不會透過Matchmaker分配房間再把ip回傳給client, 而是直接讓client去連資料庫matchgame的ip)
var Mode string
var GameTime = float64(0)                                             // 遊戲開始X秒
var TickTimePass = float64(BattleLOOP_MILISECS) / 1000.0              // 每幀時間流逝秒數
var MarketDivineJsonSkills [MarketDivineSkillCount]gameJson.JsonSkill // 本局遊戲可購買的神祉技能清單
var MyGameState = GameState_Initializing                              // 遊戲狀態

func InitGame() {
	var err error
	MarketDivineJsonSkills, err = GetRndBribeSkills()
	if err != nil {
		log.Errorf("%s InitGame: %v", logger.LOG_Game, err)
		return
	}
}
func GetRndBribeSkills() ([MarketDivineSkillCount]gameJson.JsonSkill, error) {
	allJsonSkills, err := gameJson.GetJsonSkills("Divine")
	if err != nil {
		return [MarketDivineSkillCount]gameJson.JsonSkill{}, fmt.Errorf("gameJson.GetJsonSkills()錯誤: %v", err)
	}
	var JsonSkills [MarketDivineSkillCount]gameJson.JsonSkill
	rndJsonSkills, err := utility.GetRandomNumberOfTFromMap(allJsonSkills, MarketDivineSkillCount)
	if err != nil {
		return [MarketDivineSkillCount]gameJson.JsonSkill{}, fmt.Errorf("utility.GetRandomNumberOfTFromMap錯誤: %v", err)
	}
	for i, _ := range rndJsonSkills {
		JsonSkills[i] = rndJsonSkills[i]
	}
	return JsonSkills, nil
}

type ConnectionTCP struct {
	Conn       net.Conn      // TCP連線
	Encoder    *json.Encoder // 連線編碼
	Decoder    *json.Decoder // 連線解碼
	MyLoopChan *LoopChan
}

// 關閉PackReadStopChan通道
func (loopChan *LoopChan) ClosePackReadStopChan() {
	loopChan.ChanCloseOnce.Do(func() {
		close(loopChan.StopChan)
	})
}

type LoopChan struct {
	StopChan      chan struct{} // 讀取封包Chan
	ChanCloseOnce sync.Once
}

type ConnectionUDP struct {
	Conn       net.PacketConn // UDP連線
	Addr       net.Addr       // 玩家連線地址
	ConnToken  string         // 驗證Token
	MyLoopChan *LoopChan
}

// StartFighting 開始戰鬥
func StartFighting() {
	ChangeGameState(GameState_Fighting)
	GameTime = 0
	pack := packet.Pack{
		CMD:     packet.STARTFIGHTING_TOCLIENT,
		PackID:  -1,
		Content: &packet.StartFighting_ToClient{},
	}
	MyRoom.BroadCastPacket(-1, pack)
	log.Infof("戰鬥開始")
}

// ResetGame 重置遊戲
func ResetGame() {
	ChangeGameState(GameState_Inited)
	GameTime = 0
	MyRoom.ResetRoom()
}

// 改變遊戲階段
func ChangeGameState(state GameState) {
	MyGameState = state
	log.Infof("改變遊戲狀態為: %v", MyGameState)
}

// 遊戲計時器
func RunGameTimer(stop chan struct{}) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("%s RunGameTimer錯誤: %v", logger.LOG_Room, err)
			stop <- struct{}{}
		}
	}()
	battleTicker := time.NewTicker(time.Duration(BattleLOOP_MILISECS) * time.Millisecond)
	gameTicker := time.NewTicker(time.Duration(GameLOOP_MILISECS) * time.Millisecond)
	defer battleTicker.Stop()
	defer gameTicker.Stop()

	for {
		select {
		case <-battleTicker.C:
			if MyGameState == GameState_Fighting {
				timePass()
				snedBattleStatePackToClient()
			}
		case <-gameTicker.C:
			MyRoom.KickTimeoutPlayer()
		case <-stop:
			return
		}
	}
}

func snedBattleStatePackToClient() {
	for _, v := range MyRoom.Gamers {
		if player, ok := v.(*Player); ok {
			pack := packet.Pack{
				CMD:    packet.BATTLESTATE_TOCLIENT,
				PackID: -1,
				Content: &packet.BattleState_ToClient{
					MyPlayerState:       player.GetPackPlayerState(true),
					OpponentPlayerState: player.GetOpponentPackPlayerState(),
					GameTime:            utility.RoundToDecimal(GameTime, 3),
				},
			}
			player.SendPacketToPlayer(pack)
		}
	}
}

func timePass() {
	GameTime += TickTimePass
	// 雙方觸發狀態效果
	gladiatorsTriggerBuffers()
	// 雙方移動
	gladiatorsMove()
	// 有碰撞就進行肉搏
	if checkCollision() {
		melee()
	}
}

// gladiatorsTriggerBuffers 雙方觸發狀態效果
func gladiatorsTriggerBuffers() {
	for _, v := range MyRoom.Gamers {
		if v == nil {
			continue
		}
		g := v.GetGladiator()
		if g != nil {
			g.TriggerBuffer_Time()
		}
	}
}

// gladiatorsMove 雙方移動
func gladiatorsMove() {
	for _, v := range MyRoom.Gamers {
		if v == nil {
			continue
		}
		g := v.GetGladiator()
		if g != nil {
			g.Move()
		}
	}
}

// checkCollision 碰撞檢測
func checkCollision() bool {
	if MyRoom.Gamers[0] == nil && MyRoom.Gamers[0].GetGladiator() == nil && MyRoom.Gamers[1] == nil && MyRoom.Gamers[1].GetGladiator() == nil {
		return false
	}
	dis := math.Abs(MyRoom.Gamers[0].GetGladiator().CurPos - MyRoom.Gamers[1].GetGladiator().CurPos)
	// log.Infof("pos1: %v  pos2: %v dis: %v", MyRoom.Gamers[0].GetGladiator().CurPos, MyRoom.Gamers[1].GetGladiator().CurPos, dis)
	return dis <= CollisionDis
}

// melee 雙方進行肉搏
func melee() {
	if MyRoom.Gamers[0] == nil && MyRoom.Gamers[0].GetGladiator() == nil && MyRoom.Gamers[1] == nil && MyRoom.Gamers[1].GetGladiator() == nil {
		return
	}
	var err error
	g1 := MyRoom.Gamers[0].GetGladiator()
	g2 := MyRoom.Gamers[1].GetGladiator()
	// 初始化雙方肉搏技能
	g1SpellInit := g1.GetInit()
	var g1Skill *Skill
	if g1.ActivedMeleeJsonSkill != nil {
		g1Skill, err = NewSkill(g1, g2, *g1.ActivedMeleeJsonSkill)
		if err != nil {
			log.Errorf("NewSkill錯誤")
		}
		g1SpellInit += g1Skill.JsonSkill.Init
	}
	g2SpellInit := g2.GetInit()
	var g2Skill *Skill
	if g2.ActivedMeleeJsonSkill != nil {
		g2Skill, err = NewSkill(g2, g1, *g2.ActivedMeleeJsonSkill)
		if err != nil {
			log.Errorf("NewSkill錯誤")
		}
		g2SpellInit += g2Skill.JsonSkill.Init
	}
	// 雙方技能施放
	if g1SpellInit > g2SpellInit { // g1先攻
		g1.Spell(g1Skill)
		g2.Spell(g2Skill)
	} else if g1SpellInit < g2SpellInit { // g2先攻
		g2.Spell(g2Skill)
		g1.Spell(g1Skill)
	} else { // 先攻值一樣的話就隨機一方先攻
		if utility.GetProbResult(0.5) {
			g1.Spell(g1Skill)
			g2.Spell(g2Skill)
		} else {
			g2.Spell(g2Skill)
			g1.Spell(g1Skill)
		}
	}
	// 紀錄擊退前的位置
	g1AttackPos := g1.CurPos
	g2AttackPos := g2.CurPos
	// 雙方擊退
	g1SkillKnockback := 0.0
	g2SkillKnockback := 0.0
	if g1Skill != nil {
		g1SkillKnockback += g1Skill.JsonSkill.Knockback
	}
	if g2Skill != nil {
		g2SkillKnockback += g2Skill.JsonSkill.Knockback
	}
	g1Knockback := g1.GetKnockback() + g1SkillKnockback
	g2Knockback := g2.GetKnockback() + g2SkillKnockback + 5
	g1.DoKnockback(g2Knockback)
	g2.DoKnockback(g1Knockback)
	// 雙方暈眩
	dizzyTriggerAt := GameTime + 1
	g1KnockDizzy := &Effect{
		Type:          gameJson.Dizzy,
		Duration:      2,
		Target:        g1,
		NextTriggerAt: dizzyTriggerAt,
	}
	g1.AddEffect(g1KnockDizzy)
	g2KnockDizzy := &Effect{
		Type:          gameJson.Dizzy,
		Duration:      2,
		Target:        g2,
		NextTriggerAt: dizzyTriggerAt,
	}
	g2.AddEffect(g2KnockDizzy)

	// 送Melee封包給Client
	g1SkillID := 0
	g2SkillID := 0
	if g1Skill != nil {
		g1SkillID = g1Skill.JsonSkill.ID
	}
	if g2Skill != nil {
		g2SkillID = g2Skill.JsonSkill.ID
	}
	if player, ok := MyRoom.Gamers[0].(*Player); ok {
		packMelee := packet.Pack{
			CMD: packet.MELEE_TOCLIENT,
			Content: &packet.Melee_ToClient{
				MyPlayerState:       player.GetPackPlayerState(true),
				OpponentPlayerState: player.GetOpponentPackPlayerState(),
				GameTime:            utility.RoundToDecimal(GameTime, 3),
				MyAttack: packet.PackAttack{
					Knockback: g1Knockback,
					SkillID:   g1SkillID,
					AttackPos: g1AttackPos,
				},
				OpponentAttack: packet.PackAttack{
					Knockback: g2Knockback,
					SkillID:   g2SkillID,
					AttackPos: g2AttackPos,
				},
			},
		}
		player.SendPacketToPlayer(packMelee)
	}
	if player, ok := MyRoom.Gamers[1].(*Player); ok {
		packMelee := packet.Pack{
			CMD: packet.MELEE_TOCLIENT,
			Content: &packet.Melee_ToClient{
				MyPlayerState:       player.GetPackPlayerState(true),
				OpponentPlayerState: player.GetOpponentPackPlayerState(),
				GameTime:            utility.RoundToDecimal(GameTime, 3),
				MyAttack: packet.PackAttack{
					Knockback: g2Knockback,
					SkillID:   g2SkillID,
				},
				OpponentAttack: packet.PackAttack{
					Knockback: g1Knockback,
					SkillID:   g1SkillID,
				},
			},
		}
		player.SendPacketToPlayer(packMelee)
	}

}
