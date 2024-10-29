package game

import (
	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"time"
)

type Room struct {
	gameServer    *agonesv1.GameServer
	dbMapID       string     // DB地圖ID
	dbMatchgameID string     // 就是RoomName由Lobby產生，格視為[玩家ID]_[累加數字]_[日期時間]
	matchType     string     // 配對類型
	players       []*Player  // 房間內的玩家
	creater       *Player    // 開房者
	createTime    *time.Time // 開房時間
}
