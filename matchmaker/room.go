package main

import (
	"encoding/json"
	"fmt"
	mongo "gladiatorsGoModule/mongo"
	"gladiatorsGoModule/redis"
	"gladiatorsGoModule/setting"
	"gladiatorsGoModule/utility"
	logger "matchmaker/logger"
	"matchmaker/packet"
	mSetting "matchmaker/setting"
	"net"
	"sync"
	"sync/atomic"
	"time"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	log "github.com/sirupsen/logrus"
)

type RoomReceptionist struct {
	quickRoomUshers map[string]*Usher // Key值為mapID(不同地圖有不同mapID，用來區分不同房間的玩家不會彼此配對到)
}
type Usher struct {
	roomLock        sync.RWMutex
	rooms           []*room // 已建立的房間
	lastJoinRoomIdx int     // 上一次加房索引，記錄此值避免每次找房間都是從第一間開始找
}
type room struct {
	gameServer    *agonesv1.GameServer
	dbMapID       string        // DB地圖ID
	dbMatchgameID string        // 就是RoomName由Matchmaker產生，格視為[玩家ID]_[累加數字]_[日期時間]
	matchType     string        // 配對類型
	players       []*roomPlayer // 房間內的玩家
	creater       *roomPlayer   // 開房者
	createTime    *time.Time    // 開房時間
}
type roomPlayer struct {
	id      string        // 玩家ID
	isAuth  bool          // 是否經過帳戶驗證了
	connTCP ConnectionTCP // TCP連線
	dbMapID string        // 地圖ID
	room    *room         // 房間資料
}
type ConnectionTCP struct {
	Conn    net.Conn      // TCP連線
	Encoder *json.Encoder // 連線編碼
	Decoder *json.Decoder // 連線解碼
}

func (rr *RoomReceptionist) Init() {
	rr.quickRoomUshers = make(map[string]*Usher)
	//go rr.RoutineCheckOccupiedRoom()
}

// func (rr *RoomReceptionist) RoutineCheckOccupiedRoom() {
// 	timer := time.NewTicker(ROUTINE_CHECK_OCCUPIED_ROOM * time.Minute)
// 	for {
// 		for _, usher := range rr.quickRoomUshers {
// 			usher.CheckOccupiedRoom()
// 		}
// 		<-timer.C
// 	}
// }
// func (u *Usher) CheckOccupiedRoom() {

// }
func (r *room) clearRoom() {

	log.WithFields(log.Fields{
		"players": r.players,
	}).Infof("%s ClearRoom", logger.LOG_Room)
	// 清除房間
	for i, v := range r.players {
		if v == nil {
			continue
		}
		r.players[i].LeaveRoom()
	}
	r.players = nil
	r.createTime = nil
}

// 玩家離開房間
func (p *roomPlayer) LeaveRoom() {
	p.room = nil
}

func (r *RoomReceptionist) getUsher(dbMapID string) *Usher {
	usher, ok := r.quickRoomUshers[dbMapID]
	if !ok {
		newUsher := Usher{}
		r.quickRoomUshers[dbMapID] = &newUsher
		usher = r.quickRoomUshers[dbMapID]
	}
	return usher
}

// 加入房間-快速房, 回傳房間與是否為新開房間
func (r *RoomReceptionist) JoinRoom(packID int, dbMap mongo.DBMap, player *roomPlayer) (*room, bool) {

	// 取得房間接待員
	usher := r.getUsher(dbMap.ID)

	// 找空房間
	for i, _ := range usher.rooms {
		roomIdx := (usher.lastJoinRoomIdx + i) % len(usher.rooms)
		room := usher.rooms[roomIdx]
		joined := room.AddPlayer(player)
		// 房間不可加入就換下一間檢查
		if !joined {
			usher.lastJoinRoomIdx = roomIdx
			continue
		}

		// 確認目Matchgame Server是否還活著, 掛掉就繼續找下一個
		err := CheckGameServer(room.dbMatchgameID)
		if err != nil {
			utility.RemoveFromSliceByIdx(usher.rooms, roomIdx) // 移除掛掉的房間
			log.Errorf("%s 目標遊戲房已掛: %v", logger.LOG_Room, err)
			continue
		}

		// DBMatchgame加入玩家資料(加入已存在房間時, DBMatchgame的玩家加入是在Matchmaker寫入, 但開房是在DBMatchgame寫入)
		var dbMatchgame mongo.DBMatchgame
		getDBMatchgameErr := mongo.GetDocByID(mongo.ColName.Matchgame, room.dbMatchgameID, &dbMatchgame)
		if getDBMatchgameErr != nil {
			log.Errorf("%s 取得DB資料mongo.GetDocByID(mongo.ColName.Matchgame, room.dbMatchgameID, dbMatchgame)錯誤: %v", logger.LOG_Room, getDBMatchgameErr)
			continue
		}
		joinPlayerErr := (&dbMatchgame).JoinPlayer(player.id)
		if joinPlayerErr != nil {
			log.Errorf("%s DBMatchgame加入玩家錯誤: %v", logger.LOG_Room, joinPlayerErr)
			continue
		}
		log.Infof("%s 開始更新Matchgame到DB dbMatchgame: %v", logger.LOG_Room, dbMatchgame)
		_, updateDBMatchgame := mongo.UpdateDocByInterface(mongo.ColName.Matchgame, dbMatchgame.ID, dbMatchgame)
		if updateDBMatchgame != nil {
			log.Errorf("%s mongo.UpdateDocByID(mongo.ColName.Matchgame, dbMatchgame.ID, updateData)發生錯誤", logger.LOG_Room)
		}
		log.Infof("%s 更新Matchgame到DB完成", logger.LOG_Room)

		log.WithFields(log.Fields{
			"playerID":  player.id,
			"dbMapID":   dbMap.ID,
			"roomIdx":   roomIdx,
			"room":      room,
			"dbMapData": dbMap,
		}).Infof("%s Player join an exist room", logger.LOG_Room)

		log.Infof("%s 玩家 %s 加入房間(%v/%v) 房間資料: %+v", logger.LOG_Room, player.id, room.PlayerCount(), setting.PLAYER_NUMBER, room)
		return room, false
	}

	log.Infof("%s 玩家 %s 找不到可加入的房間, 創建一個新房間(%v/%v): %+v", logger.LOG_Room, player.id, 1, setting.PLAYER_NUMBER, dbMap)
	// 找不到可加入的房間就創一個新房間
	newCreateTime := time.Now()
	newRoom := room{
		dbMapID:    dbMap.ID,
		matchType:  dbMap.MatchType,
		players:    nil,
		creater:    nil,
		createTime: &newCreateTime,
	}
	// 設定玩家所在地圖
	player.dbMapID = dbMap.ID
	// 設定玩家為開房者
	newRoom.creater = player
	// 開房者加入此新房
	newRoom.AddPlayer(player)
	// 將新房加到房間清單中
	usher.rooms = append(usher.rooms, &newRoom)
	roomIdx := len(usher.rooms) - 1
	usher.lastJoinRoomIdx = roomIdx

	// 建立遊戲(Matchgame Server)
	err := player.room.CreateGame(packID)
	if err != nil {
		log.Errorf("%s 建立Matchgame server失敗: %v", logger.LOG_Room, err)
		return nil, false
	}

	log.WithFields(log.Fields{
		"playerID":   player.id,
		"dbMapID":    dbMap.ID,
		"roomIdx":    roomIdx,
		"room":       newRoom,
		"dbRoomData": dbMap,
	}).Infof("%s Player create a new room", logger.LOG_Room)

	return &newRoom, true

}

// 訂閱Redis房間訊息
func (r *room) SubRoomMsg() {
	channelName := "Game-" + r.dbMatchgameID
	log.Infof("%s 訂閱Redis房間(%s)", logger.LOG_Room, channelName)
	msgChan := make(chan interface{})
	err := redis.Subscribe(channelName, msgChan)
	if err != nil {
		log.Errorf("%s 訂閱錯誤: %s", logger.LOG_Room, err)
		return
	}

	for msg := range msgChan {
		var data redis.RedisPubSubPack
		byteMsg := []byte(msg.(string))
		err := json.Unmarshal(byteMsg, &data)
		if err != nil {
			log.Errorf("%s JSON解析錯誤: %s", logger.LOG_Room, err)
			continue
		}

		switch data.CMD {
		case redis.CMD_PLAYERLEFT: // 玩家離開
			var playerLeftData redis.PlayerLeft
			err := json.Unmarshal(data.Content, &playerLeftData)
			if err != nil {
				log.Errorf("%s SubRoomMsg JSON 解析 Content(%s) 錯誤: %v", logger.LOG_Room, data.CMD, err)
				continue
			}
			r.RemovePlayer(playerLeftData.PlayerID) // 將該玩家從房間中移除
			log.Printf("%s 收到Matchgame玩家離開: %s", logger.LOG_Room, playerLeftData.PlayerID)
		case redis.CMD_GAMECREATED: // 房間建立
			var gameCreated redis.GameCreated
			err := json.Unmarshal(data.Content, &gameCreated)
			if err != nil {
				log.Errorf("%s SubRoomMsg JSON 解析 Content(%s) 錯誤: %v", logger.LOG_Room, data.CMD, err)
				continue
			}
			log.Printf("%s 收到Matchgame房間建立完成: %s", logger.LOG_Room, gameCreated.MatchgameID)
			creater := r.players[0]
			if creater == nil {
				return
			}
			packErr := packet.SendPack(creater.connTCP.Encoder, &packet.Pack{
				CMD:    packet.CREATEROOM_TOCLIENT,
				PackID: gameCreated.PackID,
				Content: &packet.CreateRoom_ToClient{
					CreaterID:     creater.id,
					PlayerIDs:     r.GetPlayerIDs(),
					DBMapID:       r.dbMapID,
					DBMatchgameID: r.dbMatchgameID,
					IP:            r.gameServer.Status.Address,
					Port:          r.gameServer.Status.Ports[0].Port,
					PodName:       r.gameServer.ObjectMeta.Name,
				},
			})
			if packErr != nil {
				return
			}

		}
	}
}

// 檢查此房間是否已經存在該玩家ID
func (r *room) IsIDExist(playerID string) bool {
	for _, v := range r.players {
		if v == nil {
			continue
		}
		if v.id == playerID {
			return true
		}
	}
	return false
}

// 將玩家加入此房間中
func (r *room) AddPlayer(player *roomPlayer) bool {
	// 滿足以下條件之一的房間不可加入
	// 1. 該玩家已在此房間
	// 2. 房間已滿
	if r.IsIDExist(player.id) || len(r.players) >= setting.PLAYER_NUMBER {
		return false
	}
	player.room = r                       // 將玩家的房間設定為此房間
	r.players = append(r.players, player) // 將房間的玩家清單加入此玩家
	return true
}

// 將玩家從房間中移除
func (r *room) RemovePlayer(playerID string) {
	for i, v := range r.players {
		if v != nil && v.id == playerID {
			r.players[i] = nil
			break
		}
	}
}

// 建立遊戲
func (r *room) CreateGame(packID int) error {
	var err error
	if r == nil {
		err = fmt.Errorf("%s CreateGame Room的r為nil", logger.LOG_Room)
		return err
	}
	var createGameOK bool

	// 產生房間名稱
	roomName, getRoomNameOK := r.generateRoomName()
	if !getRoomNameOK {
		createGameOK = false

		log.WithFields(log.Fields{
			"room": r,
		}).Errorf("%s Generate Room Name Failed", logger.LOG_Room)
		err = fmt.Errorf("%s Generate Room Name Failed", logger.LOG_Room)
		return err
	}

	r.dbMatchgameID = roomName
	log.WithFields(log.Fields{
		"room":     r,
		"roomName": roomName,
	}).Infof("%s Generate Room Name \n", logger.LOG_Room)

	// 建立遊戲房
	retryTimes := 0
	timer := time.NewTicker(mSetting.RETRY_INTERVAL_SECONDS * time.Second)
	for i := 0; i < mSetting.RETRY_CREATE_GAMESERVER_TIMES; i++ {
		retryTimes = i
		r.gameServer, err = CreateGameServer(packID, roomName, r.GetPlayerIDs(), r.creater.id, r.dbMapID, SelfPodName)
		if err == nil {
			createGameOK = true
			break
		}
		log.Errorf("%s CreateGameServer第%v次失敗: %v", logger.LOG_Room, i, err)
		<-timer.C
	}
	timer.Stop()

	// 寫入建立遊戲房結果Log
	if createGameOK {
		if retryTimes > 0 {
			log.WithFields(log.Fields{
				"retryTimes": retryTimes,
				"error:":     err.Error(),
			}).Infof("%s Create gameServer with retry: \n", logger.LOG_Room)
		}
	} else {
		log.WithFields(log.Fields{
			"retryTimes": mSetting.RETRY_CREATE_GAMESERVER_TIMES,
			"error:":     err.Error(),
		}).Errorf("%s Create gameServer error: \n", logger.LOG_Room)
		err = fmt.Errorf("%s Gameserver allocated failed", logger.LOG_Room)
	}

	go r.SubRoomMsg() // 訂閱房間資訊

	return err
}

var counter int64 // 房間名命名計數器
// 以創房者的id來產生房名
func (r *room) generateRoomName() (string, bool) {
	var roomName string
	if r.creater == nil {
		log.Println("Generating room name failed, creater is nil")
		return roomName, false
	}
	newCounterValue := atomic.AddInt64(&counter, 1)
	roomName = fmt.Sprintf("%s_%d_%s", r.creater.id, newCounterValue, time.Now().Format("20060102T150405"))
	return roomName, true
}

// 取遊戲房中玩家ID清單 如果該位置沒有玩家會存空字串
func (r *room) GetPlayerIDs() []string {
	ids := []string{}
	for _, v := range r.players {
		if v == nil {
			ids = append(ids, "")
			continue
		}
		ids = append(ids, v.id)
	}
	return ids
}

// 取得房間玩家數
func (r *room) PlayerCount() int {
	count := 0
	for _, v := range r.players {
		if v == nil {
			continue
		}
		count++
	}
	return count
}
