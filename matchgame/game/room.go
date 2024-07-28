package game

import (
	"fmt"
	mongo "gladiatorsGoModule/mongo"
	"gladiatorsGoModule/setting"

	// "matchgame/agones"
	logger "matchgame/logger"
	"matchgame/packet"
	"net"
	"sync"
	"time"

	// agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type Room struct {
	Gamers      [setting.PLAYER_NUMBER]Gamer // 玩家map
	RoomName    string                       // 房間名稱(也是DB文件ID)(房主UID+時間轉 MD5)
	DBMatchgame *mongo.DBMatchgame           // DB遊戲房資料
	MutexLock   sync.RWMutex
}

var MyRoom *Room // 房間

func InitGameRoom(dbMapID string, playerIDs [setting.PLAYER_NUMBER]string, roomName string, ip string, port int, podName string, nodeName string, matchmakerPodName string, roomCreatedChan chan struct{}) {
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
		Gamers:      [setting.PLAYER_NUMBER]Gamer{},
		RoomName:    roomName,
		DBMatchgame: &dbMatchgame,
	}
	MyRoom.UpdateMatchgameToDB()
	log.Infof("%s InitGameRoom完成", logger.LOG_Room)
	roomCreatedChan <- struct{}{}
}

func (r *Room) KickTimeoutPlayer() {
	for _, gamer := range r.Gamers {
		if player, _ := gamer.(*Player); player != nil {

			nowTime := time.Now()
			// 玩家無心跳超過X秒就踢出遊戲房
			// log.Infof("%s 目前玩家 %s 已經無回應 %.0f 秒了", logger.LOG_Room, player.GetID(), nowTime.Sub(player.LastUpdateAt).Seconds())
			if nowTime.Sub(player.LastUpdateAt) > time.Duration(KICK_PLAYER_SECS)*time.Second {
				ResetGame()
			}
		}

	}
}

// 傳入玩家ID取得Player
func (r *Room) GetGamerByID(gamerID string) Gamer {
	for _, gamer := range r.Gamers {
		if gamer != nil {
			if gamer.GetID() == gamerID {
				return gamer
			}
		}
	}
	return nil
}

func (r *Room) GamerExist(gamerID string) bool {
	return r.GetGamerByID(gamerID) != nil
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
func (r *Room) JoinGamer(gamer Gamer) error {
	if gamer == nil {
		return fmt.Errorf("JoinGamer傳入nil Gamer")
	}
	log.Infof("%s 玩家(%s) 嘗試加入房間 DBMatchgame: %+v", logger.LOG_Room, gamer.GetID(), r.DBMatchgame)

	r.MutexLock.Lock()
	defer r.MutexLock.Unlock()

	gamerExist := r.GamerExist(gamer.GetID())
	if gamerExist { // 斷線重連

	} else { // 玩家加入
		joinErr := r.DBMatchgame.JoinPlayer(gamer.GetID())
		if joinErr != nil {
			return fmt.Errorf("JoinPlayer時 r.DBMatchgame.JoinPlayer(gamer.GetID())錯誤: %v", joinErr)
		}
		joinIdx := -1
		for i, v := range r.Gamers {
			if v == nil {
				joinIdx = i
				break
			}
		}
		gamer.SetIdx(joinIdx)
		r.Gamers[joinIdx] = gamer
	}

	if r.GamerCount() > setting.PLAYER_NUMBER {
		return fmt.Errorf("JoinGamer玩家人數超過上限 玩家人數: %v", r.GamerCount())
	}

	r.UpdateMatchgameToDB() // 更新DB
	r.OnRoomPlayerChange()
	
	log.Infof("%s 玩家(%s) 已加入房間(%v/%v) 房間資訊: %+v", logger.LOG_Room, gamer.GetID(), r.GamerCount(), setting.PLAYER_NUMBER, r)
	return nil
}

// 重置房間
func (r *Room) ResetRoom() {
	for _, v := range r.Gamers {
		if player, ok := v.(*Player); ok {
			r.KickPlayer(player, "重置房間")
		} else if bot, ok := v.(*Bot); ok {
			r.KickBot(bot, "重置房間")
		}
	}
}

// 將玩家踢出房間
func (r *Room) KickPlayer(player *Player, reason string) {

	log.Infof("%s 嘗試踢出玩家(%s) 原因: %s", logger.LOG_Room, player.GetID(), reason)
	gamer := r.GetGamerByID(player.ID)
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
	r.Gamers[player.Idx] = nil
	r.DBMatchgame.KickPlayer(player.GetID())
	r.UpdateMatchgameToDB() // 更新房間DB

	player.CloseConnection() // 關閉連線
	r.OnRoomPlayerChange()
	log.Infof("%s 踢出Player完成, 目前Gamer人數: %v", logger.LOG_Room, r.GamerCount())
}

// 將Bot踢出房間
func (r *Room) KickBot(bot *Bot, reason string) {

	log.Infof("%s 嘗試踢出Bot(%s) 原因: %s", logger.LOG_Room, bot.GetID(), reason)
	gamer := r.GetGamerByID(bot.ID)
	if gamer == nil {
		log.Infof("%s 要踢掉的Bot已經不存在", logger.LOG_Room)
		return
	}

	r.MutexLock.Lock()
	defer r.MutexLock.Unlock()

	r.Gamers[bot.Idx] = nil
	r.DBMatchgame.KickPlayer(bot.GetID())
	r.UpdateMatchgameToDB() // 更新房間DB

	r.OnRoomPlayerChange()
	log.Infof("%s 踢出Bot完成, 目前Gamer人數: %v", logger.LOG_Room, r.GamerCount())
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

// 透過TCPConn取得玩家ID
func (r *Room) GetPlayerIDByTCPConn(conn net.Conn) string {
	for _, v := range r.Gamers {
		if player, ok := v.(*Player); ok {
			if player.ConnTCP == nil {
				continue
			}

			if player.ConnTCP.Conn == conn {
				return v.GetID()
			}
		}
	}
	return ""
}

// 透過ConnToken取得玩家座位索引
func (r *Room) GetPlayerIdByConnToken(connToken string) string {
	for _, v := range r.Gamers {
		if player, ok := v.(*Player); ok {
			if player.ConnUDP == nil {
				continue
			}

			if player.ConnUDP.ConnToken == connToken {
				return v.GetID()
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

// 送封包給遊戲房間內所有玩家(TCP), 除了指定索引(exceptPlayerIdx)的玩家, 如果要所有玩家就傳入-1就可以
func (r *Room) BroadCastPacket(exceptPlayerIdx int, pack packet.Pack) {
	// if pack.CMD != packet.SPAWN_TOCLIENT {
	// 	log.Infof("廣播封包給其他玩家 CMD: %v", pack.CMD)
	// }
	// 送封包給所有房間中的玩家
	for idx, gamer := range r.Gamers {
		if idx == exceptPlayerIdx {
			continue
		}
		if player, ok := gamer.(*Player); ok {
			player.SendPacketToPlayer(pack)
		}
	}
}

// 送封包給玩家(TCP)
func (r *Room) SendPacketToPlayer(playerID string, pack packet.Pack) {
	gamer := r.GetGamerByID(playerID)
	if player, _ := gamer.(*Player); player != nil {
		player.SendPacketToPlayer(pack)
	}
}

// 取得玩家準備狀態, 都準備好就會回傳都是true的array
func (r *Room) GetPlayerReadies() [setting.PLAYER_NUMBER]bool {
	var playerReadies [setting.PLAYER_NUMBER]bool
	idx := 0
	for _, gamer := range r.Gamers {
		if gamer == nil {
			playerReadies[idx] = false
		}
		playerReadies[idx] = gamer.IsReady()
		idx++
	}
	return playerReadies
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

// 送封包給遊戲房間內所有玩家(UDP), 除了指定Idx(exceptPlayerIdx)的玩家, 如果要所有玩家就傳入-1就可以
func (r *Room) BroadCastPacket_UDP(exceptPlayerIdx int, sendData []byte) {
	if sendData == nil {
		return
	}
	for idx, gamer := range r.Gamers {
		if exceptPlayerIdx == idx {
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

// 將房間資料更新上DB
func (room *Room) UpdateMatchgameToDB() {
	log.Infof("%s 開始更新Matchgame到DB: %v", logger.LOG_Room, room.DBMatchgame)

	_, err := mongo.AddOrUpdateDocByStruct(mongo.ColName.Matchgame, room.DBMatchgame.ID, room.DBMatchgame)
	if err != nil {
		log.Errorf("%s UpdateMatchgameToDB時mongo.UpdateDocByID(mongo.ColName.Matchgame, room.DBMatchgame.ID, updateData)發生錯誤 %v", logger.LOG_Room, err)
	}

	log.Infof("%s 更新Matchgame到DB完成", logger.LOG_Room)
}
