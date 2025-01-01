package game

import (
	"fmt"
	"gladiatorsGoModule/gameJson"
	"gladiatorsGoModule/utility"
	"net"
	"sync/atomic"

	"encoding/json"
	logger "matchgame/logger"
	"matchgame/packet"
	"sync"
	"time"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	log "github.com/sirupsen/logrus"
)

type GameState string // 目前遊戲狀態列舉
const (
	GAMESTATE_INITIALIZING         GameState = "GAMESTATE_INITIALIZING"         // 初始化中
	GAMESTATE_INITED               GameState = "GAMESTATE_INITED"               // 初始化完成
	GAMESTATE_READY                GameState = "GAMESTATE_READY"                // 房間待玩家配對
	GAMESTATE_WAITINGPLAYERS       GameState = "GAMESTATE_WAITINGPLAYERS"       // 等待雙方玩家入場
	GAMESTATE_SELECTINGDIVINESKILL GameState = "GAMESTATE_SELECTINGDIVINESKILL" // 選擇神祉技能
	GAMESTATE_COUNTINGDOWN         GameState = "GAMESTATE_COUNTINGDOWN"         // 戰鬥倒數開始中
	GAMESTATE_FIGHTING             GameState = "GAMESTATE_FIGHTING"             // 戰鬥中
	GAMESTATE_END                  GameState = "GAMESTATE_END"                  // 結束戰鬥
)

const (
	AGONES_HEALTH_PIN_INTERVAL_SEC         = 2   // 每X秒檢查AgonesServer是否正常運作(官方文件範例是用2秒)
	TCP_CONN_TIMEOUT_SEC                   = 120 // TCP連線逾時時間X秒
	KICK_PLAYER_SECS               float64 = 30  // 最長允許玩家無心跳X秒後踢出遊戲房
)

// 戰鬥
const (
	BattleLOOP_MILISECS        int     = 100  // 戰鬥每X毫秒循環
	GameLOOP_MILISECS          int     = 1000 // 遊戲每X毫秒循環
	MarketDivineSkillCount             = 4    // 有幾個神祉技能可以購買
	DivineSkillCount                   = 2    // 玩家可以買幾個神祉技能
	GladiatorSkillCount                = 6    // 玩家有幾個技能
	HandSkillCount                     = 4    // 玩家手牌技能, 索引0的技能是下一張牌
	DIST_MELEE                         = 4    //  相距X單位就肉搏
	DIST_BEFORE_MELEE                  = 4    //  相距X單位送表演肉搏技能
	DIST_LOCK_INSTANT                  = 6    //  相距X單位送鎖住技能
	MaxVigor                   float64 = 20   // 最大體力
	DefaultVigor               float64 = 5    // 初始體力
	SelectDivineCountDownSecs  int     = 10   // 選神祉技能倒數秒數
	FightingCountDownSecs      float64 = 6.4  // 戰鬥倒數秒數
	Knockwall_Dmg              int     = 10   // 撞牆傷害
	Knockwall_DmgDelayMiliSecs int     = 400  // Melee執行後幾毫秒才觸發撞牆傷害
)

var AgonesAllocated atomic.Bool // 是否已經分配到玩家
func SetAgonesAllocated(value bool) {
	AgonesAllocated.Store(value)
}
func GetAgonesAllocated() bool {
	return AgonesAllocated.Load()
}

// Tag 標籤
type Tag string

const (
	// 行為類
	SKILL_MELEE   Tag = "MELEE"         // 肉搏技能
	SKILL_INSTANT     = "SKILL_INSTANT" //  立即技能
	SKILL_DIVINE      = "SKILL_DIVINE"  // 神祉技能
	KNOCKBACK         = "KNOCKBACK"     // 擊退
	MOVE              = "MOVE"          // 移動
	IMMOBILE          = "IMMOBILE"      // 無法移動
	//  效果類
	PDMG          = "PDMG"
	MDMG          = "MDMG"
	TDMG          = "TDMG"
	RESTORE_HP    = "RESTORE_HP"    // 生命回復
	RESTORE_VIGOR = "RESTORE_VIGOR" // 體力回復
	//  Buffer類
	BUFF      = "BUFF"      // 正面效果
	DEBUFF    = "DEBUFF"    // 負面效果
	NEUTRAL   = "NEUTRAL"   // 中性效果
	NOTBUFFER = "NOTBUFFER" // 非Buffer
	PASSIVE   = "PASSIVE"   // 永久性Buffer

)

var LockInstantSkill bool = false            // 是否鎖住技能-距離一定單位會鎖住立即類技能，避免因立即技能與肉搏同時施放導致體力消耗問題
var IDAccumulator = utility.NewAccumulator() // 產生一個ID累加器
// Mode模式分為以下:
// standard:一般版本
// non-agones: 個人測試模式(不使用Agones服務, non-agones的連線方式不會透過Matchmaker分配房間再把ip回傳給client, 而是直接讓client去連資料庫matchgame的ip)
var Mode string
var GameTime = float64(0)                                             // 遊戲開始X秒
var TickTimePass = float64(BattleLOOP_MILISECS) / 1000.0              // 每幀時間流逝秒數
var MarketDivineJsonSkills [MarketDivineSkillCount]gameJson.JsonSkill // 本局遊戲可購買的神祉技能清單
var MyGameState = GAMESTATE_INITIALIZING                              // 遊戲狀態
var MyMeleeState MeleeState = MELEESTATE_NORMAL                       // 肉搏狀態

type MeleeState string

const (
	MELEESTATE_NORMAL      MeleeState = "MELEESTATE_NORMAL"      // 已經送肉搏
	MELEESTATE_WAITTOMELEE MeleeState = "MELEESTATE_WAITTOMELEE" //  已送MELEE表演封包給Client, 準備對撞
)

func InitGame() {
	var err error
	MarketDivineJsonSkills, err = GetRndBribeSkills()
	if err != nil {
		log.Errorf("%s InitGame: %v", logger.LOG_Game, err)
		return
	}
	ChangeGameState(GAMESTATE_INITED, false)
}
func GetRndBribeSkills() ([MarketDivineSkillCount]gameJson.JsonSkill, error) {
	allJsonSkills, err := gameJson.GetJsonSkills(gameJson.DIVINE)
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
	MyLoopChan *MyChan
}

// Close 關閉Channel
func (loopChan *MyChan) Close() {
	loopChan.ChanCloseOnce.Do(func() {
		close(loopChan.StopChan)
	})
}

type MyChan struct {
	StopChan      chan struct{} // 讀取封包Chan
	ChanCloseOnce sync.Once
}

type ConnectionUDP struct {
	Conn       net.PacketConn // UDP連線
	Addr       net.Addr       // 玩家連線地址
	ConnToken  string         // 驗證Token
	MyLoopChan *MyChan
}

// StartFighting 開始戰鬥
func StartFighting() {
	GameTime = 0
	ChangeGameState(GAMESTATE_FIGHTING, true)
	log.Infof("戰鬥開始")
}

// ResetGame 重置遊戲
func ResetGame(reason string) {
	MyMeleeState = MELEESTATE_NORMAL
	MyRoom.KickAllGamer(reason)
	ChangeGameState(GAMESTATE_READY, false)
	SetServerState(agonesv1.GameServerStateReady) // 將pod狀態標示回Ready，代表可以再次被Lobby分配
	SetAgonesAllocated(false)                     // 將AgonesAllocated設為false
}

// 改變遊戲階段
func ChangeGameState(state GameState, sendPack bool) {
	if MyGameState == state {
		return
	}
	MyGameState = state
	if sendPack && MyRoom != nil {
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

	// 等待5秒再開始跑
	time.Sleep(5 * time.Second)
	log.Infof("開始RunGameTimer")

	battleTicker := time.NewTicker(time.Duration(BattleLOOP_MILISECS) * time.Millisecond)
	gameTicker := time.NewTicker(time.Duration(GameLOOP_MILISECS) * time.Millisecond)
	defer battleTicker.Stop()
	defer gameTicker.Stop()
	battleStateAccumulator := utility.NewAccumulator()
	for {
		select {
		case <-stop:
			return
		case <-battleTicker.C:
			if MyGameState == GAMESTATE_FIGHTING {
				timePass()
				battleStatePackID := battleStateAccumulator.GetNextIdx()
				snedGladiatorStatePackToClient(battleStatePackID)
			}
		case <-gameTicker.C:
			MyRoom.KickTimeoutPlayer()
		}
	}
}

func snedGladiatorStatePackToClient(packID int64) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("%s snedGladiatorStatePackToClient 錯誤: %v", logger.LOG_Room, err)
		}
	}()
	for _, v := range MyRoom.Gamers {
		if player, ok := v.(*Player); ok {
			myGladiator := player.GetGladiator()
			if myGladiator == nil {
				continue
			}
			// 設定自己的PackGladiatorState
			myGState := packet.PackGladiatorState{
				CurPos:      myGladiator.CurPos.Round2(),
				CurSpd:      myGladiator.GetSpd(),
				CurVigor:    utility.RoundToDecimal(myGladiator.CurVigor, 2),
				Rush:        myGladiator.IsRush,
				EffectDatas: myGladiator.GetPackEffects(),
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
				CurPos:      opponentGladiator.CurPos.Round2(),
				CurSpd:      opponentGladiator.GetSpd(),
				CurVigor:    0, // 對手的體力是隱藏資訊
				Rush:        opponentGladiator.IsRush,
				EffectDatas: opponentGladiator.GetPackEffects(),
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
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("%s timePass 錯誤: %v", logger.LOG_Room, err)
		}
	}()
	GameTime += TickTimePass
	// 雙方觸發狀態效果
	gladiatorsTimePass()
	// 雙方移動
	gladiatorsMove()
	// 肉搏檢測
	checkMelee()
}

// gladiatorsTimePass 雙方觸發時間流逝效果
func gladiatorsTimePass() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("%s gladiatorsTimePass 錯誤: %v", logger.LOG_Room, err)
		}
	}()
	for _, v := range MyRoom.Gamers {
		if v == nil {
			continue
		}
		g := v.GetGladiator()
		if g == nil {
			continue
		}

		// 衝刺中不會恢復體力
		if !g.IsRush {
			g.AddVigor(TickTimePass)
		}

		// 觸發狀態
		g.TriggerBuffer_Time()
	}
}

// gladiatorsMove 雙方移動
func gladiatorsMove() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("%s gladiatorsMove 錯誤: %v", logger.LOG_Room, err)
		}
	}()
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

func sendBeforeMelee(gamer1, gamer2 Gamer, g1, g2 *Gladiator) {
	g1MeleeSkillID := 0
	if g1.ActivedMeleeJsonSkill != nil {
		g1MeleeSkillID = g1.ActivedMeleeJsonSkill.ID
	}
	g2MeleeSkillID := 0
	if g2.ActivedMeleeJsonSkill != nil {
		g2MeleeSkillID = g2.ActivedMeleeJsonSkill.ID
	}
	if p1, ok := gamer1.(*Player); ok {
		p1Pack := packet.Pack{
			CMD: packet.BEFORE_MELEE_TOCLIENT,
			Content: &packet.BeforeMeleeSkill_ToClient{
				MySkillID:       g1MeleeSkillID,
				OpponentSkillID: g2MeleeSkillID,
			},
		}
		p1.SendPacketToPlayer(p1Pack)
	}

	if p2, ok := gamer2.(*Player); ok {
		p2Pack := packet.Pack{
			CMD: packet.BEFORE_MELEE_TOCLIENT,
			Content: &packet.BeforeMeleeSkill_ToClient{
				MySkillID:       g2MeleeSkillID,
				OpponentSkillID: g1MeleeSkillID,
			},
		}
		p2.SendPacketToPlayer(p2Pack)
	}

}

// checkMelee 肉搏檢測
func checkMelee() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("%s checkMelee: %v", logger.LOG_Room, err)
		}
	}()
	gamer1 := MyRoom.Gamers[0]
	gamer2 := MyRoom.Gamers[1]
	if gamer1 == nil || gamer2 == nil {
		log.Errorf("checkMelee gamer1: %v gamer2: %v", gamer1, gamer2)
		return
	}

	g1 := gamer1.GetGladiator()
	g2 := gamer2.GetGladiator()
	if g1 == nil || g2 == nil {
		log.Errorf("checkMelee g1: %v g2: %v", g1, g2)
		return
	}

	dis := g1.CurPos.DistanceTo(g2.CurPos)
	// 肉搏前-施放表演肉搏技能
	if MyMeleeState == MELEESTATE_NORMAL {
		if dis < DIST_BEFORE_MELEE {
			// log.Errorf("before melee dis: %v", dis)
			MyMeleeState = MELEESTATE_WAITTOMELEE
			sendBeforeMelee(gamer1, gamer2, g1, g2)
		}
	} else if MyMeleeState == MELEESTATE_WAITTOMELEE { // 肉搏-施放技能
		if dis < DIST_MELEE {
			// log.Errorf("melee dis: %v", dis)
			melee(gamer1, gamer2, g1, g2)
			MyMeleeState = MELEESTATE_NORMAL
		}
	}

	// 鎖住立即技能
	lockInstantSkill := (dis < DIST_LOCK_INSTANT)
	if LockInstantSkill != lockInstantSkill { // 如果鎖住狀態改變
		LockInstantSkill = lockInstantSkill
		// 送封包
		pack := packet.Pack{
			CMD: packet.LOCK_INSTANT_TOCLIENT,
			Content: &packet.LockInstantSkill_ToClient{
				Lock: LockInstantSkill,
			},
		}
		MyRoom.BroadCastPacket(-1, pack)
	}
	// log.Infof("pos1: %v  pos2: %v dis: %v", MyRoom.Gamers[0].GetGladiator().CurPos, MyRoom.Gamers[1].GetGladiator().CurPos, dis)
}

// getDistBetweenGladiators 取得角鬥士之間的距離
func getDistBetweenGladiators() float64 {
	if MyRoom.Gamers[0] == nil || MyRoom.Gamers[0].GetGladiator() == nil ||
		MyRoom.Gamers[1] == nil || MyRoom.Gamers[1].GetGladiator() == nil {
		log.Error("getDistBetweenGladiators有玩家或角鬥士為nil")
		return 0
	}
	g1 := MyRoom.Gamers[0].GetGladiator()
	g2 := MyRoom.Gamers[1].GetGladiator()
	return g1.CurPos.DistanceTo(g2.CurPos)
}
