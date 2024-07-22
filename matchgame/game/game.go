package game

import (
	"fmt"
	"gladiatorsGoModule/gameJson"
	"gladiatorsGoModule/setting"
	"gladiatorsGoModule/utility"
	"net"

	"encoding/json"
	logger "matchgame/logger"
	"matchgame/packet"
	"runtime/debug"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	GameState_Inited GameState = iota
	GameState_WaitingPlayers
	GameState_Bribing
	GameState_Started
	GameState_End
)

const (
	GAMEUPDATE_MS                          = 1000  // 每X毫秒送UPDATEGAME_TOCLIENT封包給client(遊戲狀態更新並心跳檢測)
	PLAYERUPDATE_MS                        = 1000  // 每X毫秒送UPDATEPLAYER_TOCLIENT封包給client(玩家狀態更新)
	SCENEUPDATE_MS                         = 10000 // 每X毫秒送UPDATESCENE_TOCLIENT封包給client(場景狀態更新)
	AGONES_HEALTH_PIN_INTERVAL_SEC         = 1     // 每X秒檢查AgonesServer是否正常運作(官方文件範例是用2秒)
	TCP_CONN_TIMEOUT_SEC                   = 120   // TCP連線逾時時間X秒
	TIMELOOP_MILISECS              int     = 20    // 遊戲每X毫秒循環(client幀數更新精度)
	PERIODIC_SYNC_MOVE_TIME        int     = 1000  // 週期性同步位置
	KICK_PLAYER_SECS               float64 = 90    // 最長允許玩家無心跳X秒後踢出遊戲房
	MarketDivineSkillCount                 = 6     // 有幾個神祉技能可以購買
	DivineSkillCount                       = 2     // 玩家可以買幾個神祉技能
	GladiatorSkillCount                    = 6     // 玩家有幾個技能
	MarketBribeSkillCount                  = 6     // 有幾個賄賂技能可以購買
	BribeSkillCount                        = 2     // 玩家可以買幾個賄賂技能
	KNOCK_BACK_SECS                float64 = 1     // 擊退表演時間
	WAIT_BATTLE_START                      = 2     // BattleStart等待時間
	CollisionDis                           = 4     // 相距X單位就算碰撞
	MaxVigor                       int     = 20    // 最大體力
)

// 戰鬥
const (
	WallPos          = 20.0 // 牆壁的位置, 中心點距離牆壁的單位數, XX代表距離中心點XX單位的位置是牆壁, 也就是場地總共有2*XX單位
	InitGladiatorPos = 16.0 // 雙方角鬥士初始位置, 距離中心點XX單位的位置
)

var IDAccumulator = utility.NewAccumulator() // 產生一個ID累加器
// Mode模式分為以下:
// standard:一般版本
// non-agones: 個人測試模式(不使用Agones服務, non-agones的連線方式不會透過Matchmaker分配房間再把ip回傳給client, 而是直接讓client去連資料庫matchgame的ip)
var Mode string
var GameTime = float64(0)                                             // 遊戲開始X秒
var MarketDivineJsonSkills [MarketDivineSkillCount]gameJson.JsonSkill // 本局遊戲可購買的神祉技能清單
var LeftGamer Gamer                                                   // 左方玩家第1位玩家
var RightGamer Gamer                                                  // 右方玩家第2位玩家
var PeriodicSyncMoveTimer = PERIODIC_SYNC_MOVE_TIME

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

// 遊戲計時器
func RunGameTimer(stop chan struct{}) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("%s RoomTimer錯誤: %v.\n%s", logger.LOG_Room, err, string(debug.Stack()))
			stop <- struct{}{}
		}
	}()
	ticker := time.NewTicker(time.Duration(TIMELOOP_MILISECS) * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			MyRoom.KickTimeoutPlayer()
			if MyRoom.GameState == GameState_Started {
				TimePass()
			}
		case <-stop:
			return
		}
	}
}

func TimePass() {
	secs := float64(TIMELOOP_MILISECS) / 1000.0
	GameTime += secs
	stateStack := [][setting.PLAYER_NUMBER]packet.PackPlayerState{}
	timeStack := []float64{}
	notify := false
	defer func() {
		if notify && len(timeStack) > 0 {
			NotifyMove(stateStack, timeStack)
		}
	}()
	if PERIODIC_SYNC_MOVE_TIME > 0 {
		PeriodicSyncMoveTimer -= TIMELOOP_MILISECS
	}

	isCollide, collideState, collideTime, _ := gladiatorsMove(secs)
	if isCollide {
		notify = true
		stateStack = append(stateStack, collideState...)
		timeStack = append(timeStack, collideTime...)
		return
	}

	if PERIODIC_SYNC_MOVE_TIME > 0 && PeriodicSyncMoveTimer <= 0 {
		if LeftGamer.GetGladiator().CanMove() && RightGamer.GetGladiator().CanMove() {
			// 沒事，但週期性回傳修正
			notify = true
			stateStack = append(stateStack, MyRoom.GetPackPlayerStates())
			timeStack = append(timeStack, GameTime)
			log.Infof("PeriodicMoveNotify: GameTime(%f ,%d) End with POS(%d, %d), Speed: (%d, %d) ",
				float64(GameTime)/float64(TimeMili),
				GameTime,
				LeftGamer.GetGladiator().CurUnit,
				RightGamer.GetGladiator().CurUnit,
				RightGamer.GetGladiator().Speed,
				LeftGamer.GetGladiator().Speed,
			)
			PeriodicSyncMoveTimer = PERIODIC_SYNC_MOVE_TIME
		} else {
			// 事件發生中，週期性同步跳過一次
			PeriodicSyncMoveTimer += PERIODIC_SYNC_MOVE_TIME
		}
		return
	}
}

func gladiatorsMove(secs float64) (bool, [][setting.PLAYER_NUMBER]packet.PackPlayerState, []float64, error) {
	stateStack := [][setting.PLAYER_NUMBER]packet.PackPlayerState{}
	timeStack := []float64{}

	lMove := LeftGamer.GetGladiator().MoveUnitByTime(secs)
	rMove := RightGamer.GetGladiator().MoveUnitByTime(secs)
	if lMove || rMove {
		stateStack = append(stateStack, MyRoom.GetPackPlayerStates())
		timeStack = append(timeStack, GameTime)
	}

	// 碰撞
	if IsCollide() {
		collisionState, time := GetCollisionData()
		stateStack = append(stateStack, collisionState)
		timeStack = append(timeStack, time)
		return true, stateStack, timeStack, nil
	}

	return false, [][setting.PLAYER_NUMBER]packet.PackPlayerState{}, []float64{0}, nil
}

func NotifyMove(stateStack [][setting.PLAYER_NUMBER]packet.PackPlayerState, timeStack []float64) {
	pack := packet.Pack{
		CMD:    packet.BATTLESTATE_TOCLIENT,
		PackID: -1,
		Content: &packet.BattleState_ToClient{
			CMDContent:   nil,
			PlayerStates: stateStack,
			GameTime:     timeStack,
		},
	}
	MyRoom.BroadCastPacket(-1, pack)
}
