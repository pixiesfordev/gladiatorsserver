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
	GAMESTATE_INITIALIZING         GameState = "GAMESTATE_INITIALIZING"
	GAMESTATE_INITED                         = "GAMESTATE_INITED"
	GAMESTATE_WAITINGPLAYERS                 = "GAMESTATE_WAITINGPLAYERS"       // 等待雙方玩家入場
	GAMESTATE_SELECTINGDIVINESKILL           = "GAMESTATE_SELECTINGDIVINESKILL" // 選擇神祉技能
	GAMESTATE_COUNTINGDOWN                   = "GAMESTATE_COUNTINGDOWN"         // 戰鬥倒數開始中
	GAMESTATE_FIGHTING                       = "GAMESTATE_FIGHTING"             // 戰鬥中
	GAMESTATE_END                            = "GAMESTATE_END"                  // 結束戰鬥
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
	DefaultVigor                   float64 = 5    // 初始體力
	SelectDivineCountDownSecs      int     = 15   // 選神祉技能倒數秒數
	FightingCountDownSecs          int     = 4    // 戰鬥倒數秒數
	Knockwall_Dmg                  int     = 15   // 撞牆傷害
	Knockwall_DmgDelayMiliSecs     int     = 400  // Melee執行後幾毫秒才觸發撞牆傷害
)

// Tag 標籤
type Tag string

const (
	// 行為類
	SKILL_MELEE   Tag = "MELEE"         // 肉搏技能
	SKILL_INSTANT     = "SKILL_INSTANT" //  立即技能
	SKILL_DIVINE      = "SKILL_DIVINE"  // 神祉技能
	KNOCKBACK         = "KNOCKBACK"     // 擊退
	MOVE              = "MOVE"          // 移動
	//  效果類
	PDMG          = "PDMG"
	MDMG          = "MDMG"
	RESTORE_HP    = "RESTORE_HP"    // 生命回復
	RESTORE_VIGOR = "RESTORE_VIGOR" // 體力回復
	//  Buffer類
	BUFF      = "BUFF"      // 正面效果
	DEBUFF    = "DEBUFF"    // 負面效果
	NEUTRAL   = "NEUTRAL"   // 中性效果
	NOTBUFFER = "NOTBUFFER" // 非Buffer
	PASSIVE   = "PASSIVE"   // 永久性Buffer

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
var MyGameState = GAMESTATE_INITIALIZING                              // 遊戲狀態

func InitGame() {
	var err error
	MarketDivineJsonSkills, err = GetRndBribeSkills()
	if err != nil {
		log.Errorf("%s InitGame: %v", logger.LOG_Game, err)
		return
	}
	ChangeGameState(GAMESTATE_INITED)
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
	GameTime = 0
	ChangeGameState(GAMESTATE_FIGHTING)
	log.Infof("戰鬥開始")
}

// ResetGame 重置遊戲
func ResetGame() {
	MyRoom.ResetRoom()
}

// 改變遊戲階段
func ChangeGameState(state GameState) {
	if MyGameState == state {
		return
	}
	MyGameState = state
	log.Infof("MyRoom: %v", MyRoom)
	if MyRoom != nil {
		// 回送封包
		myPack := packet.Pack{
			CMD:    packet.GAMESTATE_TOCLIENT,
			PackID: -1,
			Content: &packet.GameState_ToClient{
				State: string(MyGameState),
			},
		}
		MyRoom.BroadCastPacket(-1, myPack)
	}
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
	battleStateAccumulator := utility.NewAccumulator()
	for {
		select {
		case <-battleTicker.C:
			if MyGameState == GAMESTATE_FIGHTING {
				timePass()
				battleStatePackID := battleStateAccumulator.GetNextIdx()
				snedGladiatorStatePackToClient(battleStatePackID)
			}
		case <-gameTicker.C:
			MyRoom.KickTimeoutPlayer()
		case <-stop:
			return
		}
	}
}

func snedGladiatorStatePackToClient(packID int64) {
	for _, v := range MyRoom.Gamers {
		if player, ok := v.(*Player); ok {
			myGladiator := player.GetGladiator()
			if myGladiator == nil {
				continue
			}
			// 設定自己的PackGladiatorState
			myGState := packet.PackGladiatorState{
				CurPos:      utility.RoundToDecimal(myGladiator.CurPos, 3),
				CurSpd:      myGladiator.GetSpd(),
				CurVigor:    myGladiator.CurVigor,
				Rush:        myGladiator.IsRush,
				EffectTypes: myGladiator.GetEffectStrs(),
			}
			// 設定對手的PackGladiatorState
			opponentGamer := player.GetOpponent()
			if opponentGamer == nil {
				log.Errorf("對手Gamer為nil")
				return
			}
			opponentGladiator := opponentGamer.GetGladiator()
			if opponentGladiator == nil {
				log.Errorf("對手opponentGladiator為nil")
				return
			}
			opponentGState := packet.PackGladiatorState{
				CurPos:      utility.RoundToDecimal(opponentGladiator.CurPos, 3),
				CurSpd:      opponentGladiator.GetSpd(),
				CurVigor:    0, // 對手的體力是隱藏資訊
				Rush:        opponentGladiator.IsRush,
				EffectTypes: opponentGladiator.GetEffectStrs(),
			}

			pack := packet.Pack{
				CMD:    packet.GLADIATORSTATES_TOCLIENT,
				PackID: packID,
				Content: &packet.GladiatorStates_ToClient{
					Time:          time.Now().UnixMilli(),
					MyState:       myGState,
					OpponentState: opponentGState,
				},
			}
			player.SendPacketToPlayer(pack)
		}
	}
}

func timePass() {
	GameTime += TickTimePass
	// 雙方觸發狀態效果
	gladiatorsTimePass()
	// 雙方移動
	gladiatorsMove()
	// 有碰撞就進行肉搏
	if checkCollision() {
		melee()
	}
}

// gladiatorsTimePass 雙方觸發時間流逝效果
func gladiatorsTimePass() {
	for _, v := range MyRoom.Gamers {
		if v == nil {
			continue
		}
		g := v.GetGladiator()
		if g == nil {
			continue
		}
		// 體力恢復
		g.AddVigor(TickTimePass)
		// 衝刺消耗體力
		if g.IsRush {
			if g.CurVigor >= TickTimePass {
				g.AddVigor(-TickTimePass)
			} else {
				g.SetRush(false)
			}
		}
		// 觸發狀態
		g.TriggerBuffer_Time()
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

// getDistBetweenGladiators 取得角鬥士之間的距離
func getDistBetweenGladiators() float64 {
	if MyRoom.Gamers[0] == nil || MyRoom.Gamers[0].GetGladiator() != nil ||
		MyRoom.Gamers[1] == nil || MyRoom.Gamers[1].GetGladiator() != nil {
		log.Error("getDistBetweenGladiators有玩家或角鬥士為nil")
		return 0
	}
	return math.Abs(MyRoom.Gamers[0].GetGladiator().CurPos - MyRoom.Gamers[1].GetGladiator().CurPos)
}
