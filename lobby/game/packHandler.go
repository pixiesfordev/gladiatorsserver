package game

import (
	"fmt"
	logger "lobby/logger"
	"lobby/packet"
	"time"
)

// 處理TCP訊息
func HandleTCPMsg(player *Player, pack packet.Pack) error {
	// 處理各類型封包
	switch pack.CMD {
	// ==========心跳==========
	case packet.PING:
		player.LastUpdateAt = time.Now() // 更新心跳
		// 回送Ping
		pack := packet.Pack{
			CMD:     packet.PING_TOCLIENT,
			PackID:  pack.PackID,
			Content: &packet.Ping_ToClient{},
		}
		player.SendPacketToPlayer(pack)
	default:
		return fmt.Errorf("%s 收到來自 %v 的未知封包: %v", logger.LOG_Main, player.ID, pack.CMD)
	}

	return nil
}
