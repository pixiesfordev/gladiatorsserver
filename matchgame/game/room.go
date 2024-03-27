package game

import (
	"errors"
	mongo "gladiatorsGoModule/mongo"
	"gladiatorsGoModule/setting"
	// "matchgame/agones"
	logger "matchgame/logger"
	"matchgame/packet"
	gSetting "matchgame/setting"
	"net"
	"runtime/debug"
	"sync"
	"time"

	// agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type GameState int // 目前遊戲狀態列舉

const (
	Init GameState = iota
	Start
	End
)

const (
	TIMELOOP_MILISECS int     = 100 // 遊戲每X毫秒循環
	KICK_PLAYER_SECS  float64 = 60  // 最長允許玩家無心跳X秒後踢出遊戲房
)

type Room struct {
	Gamers      map[string]Gamer   // 玩家map
	RoomName    string             // 房間名稱(也是DB文件ID)(房主UID+時間轉 MD5)
	GameState   GameState          // 遊戲狀態
	DBMatchgame *mongo.DBMatchgame // DB遊戲房資料
	GameTime    float64            // 遊戲開始X秒
	MutexLock   sync.RWMutex
}

var MyRoom *Room // 房間

// Mode模式分為以下:
// standard:一般版本
// non-agones: 個人測試模式(不使用Agones服務, non-agones的連線方式不會透過Matchmaker分配房間再把ip回傳給client, 而是直接讓client去連資料庫matchgame的ip)
var Mode string

func InitGameRoom(dbMapID string, playerIDs [setting.PLAYER_NUMBER]string, roomName string, ip string, port int, podName string, nodeName string, matchmakerPodName string, roomChan chan *Room) {
	log.Infof("%s InitGameRoom開始", logger.LOG_Room)
	if MyRoom != nil {
		log.Errorf("%s MyRoom已經被初始化過", logger.LOG_Room)
		return
	}

	log.Infof("%s 設定dbMatchgame資料", logger.LOG_Room)
	// 設定dbMatchgame資料
	var dbMatchgame mongo.DBMatchgame
	dbMatchgame.ID = roomName
	dbMatchgame.CreatedAt = time.Now()
	dbMatchgame.DBMapID = dbMapID
	dbMatchgame.PlayerIDs = playerIDs
	dbMatchgame.IP = ip
	dbMatchgame.Port = port
	dbMatchgame.NodeName = nodeName
	dbMatchgame.PodName = podName
	dbMatchgame.MatchmakerPodName = matchmakerPodName

	log.Infof("%s 初始化房間設定", logger.LOG_Room)
	// 初始化房間設定
	MyRoom = &Room{
		RoomName:    roomName,
		GameState:   Init,
		DBMatchgame: &dbMatchgame,
		GameTime:    0,
	}
	go RoomLoop() // 開始房間循環
	log.Infof("%s InitGameRoom完成", logger.LOG_Room)
	roomChan <- MyRoom
}

// 房間循環
func RoomLoop() {
	ticker := time.NewTicker(gSetting.ROOMLOOP_MS * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {

	}
}

// 傳入玩家ID取得Player
func (r *Room) GetGamerByID(gamerID string) Gamer {
	if gamer, ok := r.Gamers[gamerID]; ok {
		return gamer
	} else {
		return nil
	}
}

func (r *Room) GamerExist(gamerID string) bool {
	_, ok := r.Gamers[gamerID]
	return ok
}

// 取得房間玩家數
func (r *Room) GamerCount() int {
	count := 0
	for _, v := range r.Gamers {
		if v != nil {
			count++
		}
	}
	return count
}

// 把玩家加到房間中, 成功時回傳true
func (r *Room) JoinGamer(gamer Gamer) bool {
	if gamer == nil {
		log.Errorf("%s JoinGamer傳入nil Gamer", logger.LOG_Room)
		return false
	}
	log.Infof("%s 玩家(%s) 嘗試加入房間 DBMatchgame: %+v", logger.LOG_Room, gamer.GetID(), r.DBMatchgame)

	r.MutexLock.Lock()
	defer r.MutexLock.Unlock()

	gamerExist := r.GamerExist(gamer.GetID())
	if gamerExist { // 斷線重連

	} else { // 玩家加入
		joinErr := r.DBMatchgame.JoinPlayer(gamer.GetID())
		if joinErr != nil {
			log.Errorf("%s JoinPlayer時 r.DBMatchgame.JoinPlayer(gamer.GetID())錯誤: %v", logger.LOG_Room, joinErr)
			return false
		}
		r.Gamers[gamer.GetID()] = gamer
	}
	if r.GamerCount() > setting.PLAYER_NUMBER {
		log.Errorf("%s JoinGamer玩家人數超過上限 玩家人數: %v", logger.LOG_Room, r.GamerCount())
		return false
	}

	r.UpdateMatchgameToDB() // 更新DB
	r.OnRoomPlayerChange()

	log.Infof("%s 玩家(%s) 已加入房間(%v/%v) 房間資訊: %+v", logger.LOG_Room, gamer.GetID(), r.GamerCount(), setting.PLAYER_NUMBER, r)
	return true
}

// 將玩家踢出房間
func (r *Room) KickPlayer(player *Player, reason string) {

	log.Infof("%s 嘗試踢出玩家(%s) 原因: %s", logger.LOG_Room, player.GetID(), reason)
	gamer := r.GetGamerByID(player.id)
	if gamer == nil {
		log.Infof("%s 要踢掉的玩家已經不存在", logger.LOG_Room)
		return
	}

	// 取mongoDB player doc
	var mongoPlayerDoc mongo.DBPlayer
	getPlayerDocErr := mongo.GetDocByID(mongo.ColName.Player, player.GetID(), &mongoPlayerDoc)
	if getPlayerDocErr != nil {
		log.Errorf("%s 取mongoDB player doc資料發生錯誤: %v", logger.LOG_Room, getPlayerDocErr)
		return
	}

	r.MutexLock.Lock()
	defer r.MutexLock.Unlock()

	// 更新玩家DB資料
	updatePlayerBson := bson.D{
		{Key: "gold", Value: player.GetGold()}, // 玩家金幣
		{Key: "inMatchgameID", Value: ""},      // 玩家不在遊戲房內了
	}
	_, updateErr := mongo.UpdateDocByBsonD(mongo.ColName.Player, player.GetID(), updatePlayerBson) // 更新DB DBPlayer
	if updateErr != nil {
		log.Errorf("%s 更新玩家 %s DB資料錯誤: %v", logger.LOG_Room, player.GetID(), updateErr)
	} else {
		log.Infof("%s 更新玩家 %s DB資料", logger.LOG_Room, player.GetID())
	}

	r.Gamers[player.GetID()] = nil
	r.DBMatchgame.KickPlayer(player.GetID())
	r.UpdateMatchgameToDB() // 更新房間DB

	player.CloseConnection() // 關閉連線
	r.OnRoomPlayerChange()

	log.Infof("%s 踢出玩家完成", logger.LOG_Room)
}

// 房間人數有異動處理
func (r *Room) OnRoomPlayerChange() {
	if r == nil {
		return
	}
	playerCount := r.GamerCount()
	// log.Infof("%s 根據玩家數量決定是否升怪 玩家數量: %v", logger.LOG_MonsterSpawner, playerCount)

	if playerCount >= setting.PLAYER_NUMBER { // 滿房
	} else if playerCount == 0 { // 空房間處理
	} else { // 有人但沒有滿房
	}
}

// 處理TCP訊息
func (r *Room) HandleTCPMsg(conn net.Conn, pack packet.Pack) error {
	player := r.GetPlayerByTCPConn(conn)
	if player == nil {
		log.Errorf("%s HandleMessage 錯誤, 玩家不存在連線清單中", logger.LOG_Room)
		return errors.New("HandleMessage 錯誤, 玩家不存在連線清單中")
	}
	conn.SetDeadline(time.Time{}) // 移除連線超時設定
	// 處理各類型封包
	switch pack.CMD {
	// ==========賄賂==========
	case packet.BRIBE:
		// r.SendPacketToPlayer(player.Index, &packet.Pack{
		// 	CMD:    packet.UPDATESCENE_TOCLIENT,
		// 	PackID: -1,
		// 	Content: &packet.UpdateScene_ToClient{
		// 		Spawns:       r.MSpawner.Spawns,
		// 		SceneEffects: r.SceneEffects,
		// 	},
		// })
	}

	return nil
}

// 透過TCPConn取得玩家ID
func (r *Room) GetPlayerIDByTCPConn(conn net.Conn) string {
	for id, v := range r.Gamers {
		if player, ok := v.(*Player); ok {
			if player.ConnTCP == nil {
				continue
			}

			if player.ConnTCP.Conn == conn {
				return id
			}
		}
	}
	return ""
}

// 透過ConnToken取得玩家座位索引
func (r *Room) GetPlayerIdByConnToken(connToken string) string {
	for id, v := range r.Gamers {
		if player, ok := v.(*Player); ok {
			if player.ConnUDP == nil {
				continue
			}

			if player.ConnUDP.ConnToken == connToken {
				return id
			}
		}
	}
	return ""
}

// 透過TCPConn取得玩家
func (r *Room) GetPlayerByTCPConn(conn net.Conn) *Player {
	for _, v := range r.Gamers {
		if player, ok := v.(*Player); ok {
			if player.ConnTCP == nil {
				continue
			}
			if player.ConnTCP.Conn == conn {
				return player
			}
		}
	}
	return nil
}

// 透過ConnToken取得玩家
func (r *Room) GetPlayerByConnToken(connToken string) *Player {
	for _, v := range r.Gamers {
		if player, ok := v.(*Player); ok {
			if player.ConnUDP == nil {
				continue
			}
			if player.ConnUDP.ConnToken == connToken {
				return player
			}
		}
	}
	return nil
}

// 改變遊戲狀態
func (r *Room) ChangeState(state GameState) {
	r.GameState = state
}

// 送封包給遊戲房間內所有玩家(TCP), 除了指定索引(exceptPlayerIdx)的玩家, 如果要所有玩家就傳入-1就可以
func (r *Room) BroadCastPacket(exceptPlayerID string, pack *packet.Pack) {
	// if pack.CMD != packet.SPAWN_TOCLIENT {
	// 	log.Infof("廣播封包給其他玩家 CMD: %v", pack.CMD)
	// }
	// 送封包給所有房間中的玩家
	for id, gamer := range r.Gamers {
		if id == exceptPlayerID {
			continue
		}
		if player, ok := gamer.(*Player); ok {
			if player == nil || player.ConnTCP.Conn == nil {
				continue
			}
			err := packet.SendPack(player.ConnTCP.Encoder, pack)
			if err != nil {
				log.Errorf("%s 廣播封包(%s)錯誤: %v", logger.LOG_Room, pack.CMD, err)
			}
		}
	}
}

// 送封包給玩家(TCP)
func (r *Room) SendPacketToPlayer(playerID string, pack *packet.Pack) {
	gamer := r.GetGamerByID(playerID)
	if player, _ := gamer.(*Player); player != nil {
		if player.ConnTCP.Conn == nil {
			return
		}
		err := packet.SendPack(player.ConnTCP.Encoder, pack)
		if err != nil {
			log.Errorf("%s SendPacketToPlayer error: %v", logger.LOG_Room, err)
		}
	}
}

// 取得要送封包的玩家陣列
func (r *Room) GetPacketPlayers() [setting.PLAYER_NUMBER]*packet.PackPlayer {
	var players [setting.PLAYER_NUMBER]*packet.PackPlayer
	idx := 0
	for _, gamer := range r.Gamers {
		if gamer == nil {
			continue
		}
		players[idx] = &packet.PackPlayer{
			DBPlayerID:    gamer.GetID(),
			DBGladiatorID: gamer.GetGladiator().ID,
		}
	}
	return players
}

// 送封包給玩家(UDP)
func (r *Room) SendPacketToPlayer_UDP(playerID string, sendData []byte) {
	gamer := r.GetGamerByID(playerID)
	if player, _ := gamer.(*Player); player != nil {
		sendData = append(sendData, '\n')
		_, sendErr := player.ConnUDP.Conn.WriteTo(sendData, player.ConnUDP.Addr)
		if sendErr != nil {
			log.Errorf("%s (UDP)送封包錯誤 %s", logger.LOG_Room, sendErr.Error())
			return
		}
	}

}

// 送封包給遊戲房間內所有玩家(UDP), 除了指定ID(exceptPlayerID)的玩家, 如果要所有玩家就傳入-1就可以
func (r *Room) BroadCastPacket_UDP(exceptPlayerID string, sendData []byte) {
	if sendData == nil {
		return
	}
	for key, gamer := range r.Gamers {
		if exceptPlayerID == key {
			continue
		}
		if player, _ := gamer.(*Player); player != nil {
			sendData = append(sendData, '\n')
			_, sendErr := player.ConnUDP.Conn.WriteTo(sendData, player.ConnUDP.Addr)
			if sendErr != nil {
				log.Errorf("%s (UDP)送封包錯誤 %s", logger.LOG_Room, sendErr.Error())
				return
			}
		}
	}
}

// 遊戲計時器
func (r *Room) RoomTimer(stop chan struct{}) {
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
			r.GameTime += float64(TIMELOOP_MILISECS) / float64(1000) // 更新遊戲時間
			for _, gamer := range r.Gamers {
				if player, _ := gamer.(*Player); player != nil {

					nowTime := time.Now()
					// 玩家無心跳超過X秒就踢出遊戲房
					// log.Infof("%s 目前玩家 %s 已經無回應 %.0f 秒了", logger.LOG_Room, player.DBPlayer.ID, nowTime.Sub(player.LastUpdateAt).Seconds())
					if nowTime.Sub(player.LastUpdateAt) > time.Duration(KICK_PLAYER_SECS)*time.Second {
						MyRoom.KickPlayer(player, "玩家心跳逾時")
					}
				}

			}
		case <-stop:
			return
		}
	}
}

// 將房間資料更新上DB
func (room *Room) UpdateMatchgameToDB() {
	log.Infof("%s 開始更新Matchgame到DB: %v", logger.LOG_Room, room.DBMatchgame)

	_, err := mongo.UpdateDocByInterface(mongo.ColName.Matchgame, room.DBMatchgame.ID, room.DBMatchgame)
	if err != nil {
		log.Errorf("%s UpdateMatchgameToDB時mongo.UpdateDocByID(mongo.ColName.Matchgame, room.DBMatchgame.ID, updateData)發生錯誤 %v", logger.LOG_Room, err)
	}

	log.Infof("%s 更新Matchgame到DB完成", logger.LOG_Room)
}
