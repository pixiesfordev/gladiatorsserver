package game

import (
	"fmt"
	"gladiatorsGoModule/gameJson"
	"gladiatorsGoModule/utility"
	"net"

	"encoding/json"
	logger "matchgame/logger"
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
	TIMELOOP_MILISECS              int     = 100   // 遊戲每X毫秒循環
	KICK_PLAYER_SECS               float64 = 20    // 最長允許玩家無心跳X秒後踢出遊戲房
	MarketBribeSkillCount                  = 6     // 有幾個賄賂技能可以購買
	BribeSkillCount                        = 2     // 玩家可以買幾個賄賂技能
	GladiatorSkillCount                    = 6     // 玩家有幾個技能
)

// 戰鬥
const (
	WallPos          = 100 // 牆壁的位置, 中心點距離牆壁的單位數, 100代表距離中心點100單位的位置是牆壁, 也就是場地總共有200單位
	InitGladiatorPos = 50  // 雙方角鬥士初始位置, 距離中心點50單位的位置
	GridUnit         = 20  // 每X單位(unit)算1格(grid)

)

var IDAccumulator = utility.NewAccumulator() // 產生一個ID累加器
// Mode模式分為以下:
// standard:一般版本
// non-agones: 個人測試模式(不使用Agones服務, non-agones的連線方式不會透過Matchmaker分配房間再把ip回傳給client, 而是直接讓client去連資料庫matchgame的ip)
var Mode string
var GameTime = float64(0)                                      // 遊戲開始X秒
var MarketJsonBribes [MarketBribeSkillCount]gameJson.JsonBribe // 本局遊戲可購買的賄賂技能清單
var LeftGamer Gamer                                            // 左方玩家第1位玩家
var RightGamer Gamer                                           // 右方玩家第2位玩家

func InitGame() {
	var err error
	MarketJsonBribes, err = GetRndBribeSkills()
	if err != nil {
		log.Errorf("%s InitGame: %v", logger.LOG_Game, err)
		return
	}
}
func GetRndBribeSkills() ([MarketBribeSkillCount]gameJson.JsonBribe, error) {
	allJsonBribes, err := gameJson.GetJsonBribes()
	if err != nil {
		return [MarketBribeSkillCount]gameJson.JsonBribe{}, fmt.Errorf("gameJson.GetJsonSkills()錯誤: %v", err)
	}
	var jsonBribes [MarketBribeSkillCount]gameJson.JsonBribe
	rndJsonBribes, err := utility.GetRandomNumberOfTFromMap(allJsonBribes, MarketBribeSkillCount)
	if err != nil {
		return [MarketBribeSkillCount]gameJson.JsonBribe{}, fmt.Errorf("utility.GetRandomNumberOfTFromMap錯誤: %v", err)
	}
	for i, _ := range rndJsonBribes {
		jsonBribes[i] = rndJsonBribes[i]
	}
	return jsonBribes, nil
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
	GameTime += float64(TIMELOOP_MILISECS) / float64(1000) // 更新遊戲時間

	// 腳色移動
	MyRoom.Gamers[0].GetGladiator().Move()
	MyRoom.Gamers[1].GetGladiator().Move()
	log.Errorf("Left Unit: %v Grid: %v", MyRoom.Gamers[0].GetGladiator().CurUnit, MyRoom.Gamers[0].GetGladiator().CurGrid())
	log.Errorf("Right Unit: %v Grid: %v", MyRoom.Gamers[1].GetGladiator().CurUnit, MyRoom.Gamers[1].GetGladiator().CurGrid())

	// 碰撞
	if IsCollide() {
		RunCollision()
	}
}

func IsCollide() bool {
	return LeftGamer.GetGladiator().CurGrid() >= RightGamer.GetGladiator().CurGrid()
}

func RunCollision() {

}
