package game

import (
	"encoding/json"
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
	case packet.MATCH: // 配對
		content := packet.Match{}
		err := json.Unmarshal([]byte(pack.GetContentStr()), &content)
		if err != nil {
			return fmt.Errorf("%s 解析 %s 錯誤: %v", logger.LOG_Action, pack.CMD, err)
		}
		err = MyUsher.Match(player, content.DBMapID)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("%s 收到來自 %v 的未知封包: %v", logger.LOG_Main, player.ID, pack.CMD)
	}

	return nil
}
