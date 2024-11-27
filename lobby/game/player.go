package game

import (
	"encoding/json"
	"fmt"
	"lobby/logger"
	"lobby/packet"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

var Players map[string]*Player

type Player struct {
	ID            string // 玩家ID
	LastUpdateAt  time.Time
	ConnTCP       *ConnectionTCP // TCP連線
	MyRoom        *Room          // 目前所在房間資料
	QueueJoinTime time.Time      // 配房排隊時間
}

type ConnectionTCP struct {
	Conn    net.Conn      // TCP連線
	Encoder *json.Encoder // 連線編碼
	Decoder *json.Decoder // 連線解碼
}

func NewPlayer(id string, conn net.Conn) (*Player, error) {
	if _, ok := Players[id]; ok {
		return nil, fmt.Errorf("%v 玩家 %v 已經在房間中了", logger.LOG_Player, id)
	}

	player := Player{
		ID: id,
		ConnTCP: &ConnectionTCP{
			Conn:    conn,
			Encoder: json.NewEncoder(conn),
			Decoder: json.NewDecoder(conn),
		},
	}
	return &player, nil
}

// SendPacketToPlayer (TCP)送封包給玩家
func (p *Player) SendPacketToPlayer(pack packet.Pack) {
	if p.ConnTCP.Conn == nil {
		return
	}
	err := packet.SendPack(p.ConnTCP.Encoder, pack)
	if err != nil {
		log.Errorf("%s SendPacketToPlayer error: %v", logger.LOG_Player, err)
	}
}
