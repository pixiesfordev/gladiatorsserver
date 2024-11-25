package game

import (
	"fmt"
	"gladiatorsGoModule/mongo"
	logger "lobby/logger"
	"lobby/packet"
	"sync"
	"time"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	log "github.com/sirupsen/logrus"
)

type Usher struct {
	Rooms           map[string]*Room // 已建立的房間
	RoomLock        sync.RWMutex
	LastJoinRoomIdx int                  // 上一次加房索引，記錄此值避免每次找房間都是從第一間開始找
	Queue           map[string][]*Player // 排隊中的玩家
	QueueLock       sync.Mutex
}
type Room struct {
	GameServer    *agonesv1.GameServer
	DbMapID       string    // DB地圖ID
	DbMatchgameID string    // 就是RoomName由Lobby產生，格視為[DBMapID]_[玩家ID]_[時間戳]
	MatchType     string    // 配對類型
	Players       []*Player // 房間內的玩家
	Creater       *Player   // 開房者
	CreateTime    time.Time // 開房時間
}

// NewUsher 初始化 Usher
func NewUsher() *Usher {
	return &Usher{
		Rooms: make(map[string]*Room),
		Queue: make(map[string][]*Player),
	}
}

func (u *Usher) Match(player *Player, dbMapID string) error {
	dbMap, ok := GetDBMap(dbMapID)
	if !ok {
		return fmt.Errorf("%v 玩家 %s 加入地圖 %s 不存在", logger.LOG_Room, player.ID, dbMapID)
	}
	// 根據DB地圖設定來配對玩家
	switch dbMap.MatchType {
	case mongo.MATCHTYPE_QUICK: // 快速配對
		u.AddPlayerToQueue(player, dbMapID)
	default:
		errMsg := fmt.Sprintf("玩家 %v 傳入不支援的配對類型: %v", player.ID, dbMap.MatchType)
		log.Errorf(errMsg)
		pack := packet.Pack{
			CMD:    packet.MATCH_TOCLIENT,
			ErrMsg: errMsg,
		}
		player.SendPacketToPlayer(pack)
		return fmt.Errorf(errMsg)
	}
	return nil
}

// AddPlayerToQueue 將玩家加入排隊
func (u *Usher) AddPlayerToQueue(player *Player, dbMapID string) {
	u.QueueLock.Lock()
	defer u.QueueLock.Unlock()

	u.Queue[dbMapID] = append(u.Queue[dbMapID], player)
	log.Infof("%v 玩家 %s 加入地圖 %s 的排隊", logger.LOG_Room, player.ID, dbMapID)
}

// RemovePlayerFromQueue 將玩家從排隊中移除
func (u *Usher) RemovePlayerFromQueue(player *Player, mapID string) {
	u.QueueLock.Lock()
	defer u.QueueLock.Unlock()

	players := u.Queue[mapID]
	for i, p := range players {
		if p.ID == player.ID {
			u.Queue[mapID] = append(players[:i], players[i+1:]...)
			log.Infof("%v 玩家 %s 離開地圖 %s 的排隊", logger.LOG_Room, player.ID, mapID)
			return
		}
	}
}

// MatchPlayers 配對玩家
func (u *Usher) MatchPlayers() {
	for {
		log.Infof("%v 配對玩家循環", logger.LOG_Room)
		var matchGroups []struct {
			mapID   string
			players []*Player
		}

		// 在較小的臨界區內完成玩家選擇
		u.QueueLock.Lock()
		for mapID, players := range u.Queue {
			queueLen := len(players)
			if queueLen < ROOM_MAX_PLAYER {
				continue
			}

			startIdx := u.LastJoinRoomIdx % queueLen
			matchedPlayers := make([]*Player, ROOM_MAX_PLAYER)
			for i := 0; i < ROOM_MAX_PLAYER; i++ {
				matchedPlayers[i] = players[(startIdx+i)%queueLen]
			}

			u.LastJoinRoomIdx = (startIdx + ROOM_MAX_PLAYER) % queueLen
			u.Queue[mapID] = append(players[:startIdx], players[startIdx+ROOM_MAX_PLAYER:]...)

			// 將配對結果保存起來，稍後處理
			matchGroups = append(matchGroups, struct {
				mapID   string
				players []*Player
			}{mapID, matchedPlayers})
		}
		u.QueueLock.Unlock()

		// 在釋放 QueueLock 後再創建房間
		for _, group := range matchGroups {
			u.CreateRoom(group.mapID, group.players...)
		}

		time.Sleep(time.Duration(ROOM_MATCH_LOOP_MILISEC) * time.Millisecond)
	}
}

// CreateRoom 建立房間
func (u *Usher) CreateRoom(dbMapID string, players ...*Player) {
	timestamp := time.Now()
	roomID := fmt.Sprintf("%s_%v_%v", dbMapID, players[0].ID, timestamp)
	log.Infof("%v 開始建立房間 %v，玩家資料: %v", logger.LOG_Room, roomID, players)
	playerIDs := make([]string, len(players))
	for i, player := range players {
		playerIDs[i] = player.ID
	}

	// 透過Agones分配遊戲房
	timer := time.NewTicker(CREATEROOM_WAIT_SECONDS * time.Second)
	defer timer.Stop()

	room := &Room{
		DbMapID:       dbMapID,
		DbMatchgameID: roomID,
		Players:       players,
		Creater:       players[0],
		CreateTime:    timestamp,
	}

	for i := 0; i < CREATEROOM_RETRY_TIMES; i++ {
		gs, err := ApplyGameServer(roomID, playerIDs, players[0].ID, dbMapID, SelfPodName)
		if err == nil {
			room.GameServer = gs
			break
		}
		log.Errorf("%s CreateGameServer第%v次失敗: %v", logger.LOG_Room, i, err)
		if i < CREATEROOM_RETRY_TIMES-1 {
			<-timer.C
		}
	}

	u.RoomLock.Lock()
	defer u.RoomLock.Unlock()

	for _, player := range players {
		player.MyRoom = room
	}
	u.Rooms[roomID] = room

	// 送配對成功訊息給配對到的玩家
	pack := packet.Pack{
		CMD: packet.MATCH_TOCLIENT,
		Content: &packet.Match_ToClient{
			CreaterID:     players[0].ID,
			PlayerIDs:     playerIDs,
			DBMapID:       dbMapID,
			DbMatchgameID: roomID,
		},
	}
	for _, player := range players {
		player.SendPacketToPlayer(pack)
	}

	log.Infof("%v 建立房間 %s 成功，%d 位玩家配對成功", logger.LOG_Room, roomID, len(players))
}
