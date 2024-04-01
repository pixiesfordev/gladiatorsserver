package game

import (
	"encoding/json"
	"errors"
	"fmt"
	mongo "gladiatorsGoModule/mongo"
	logger "matchgame/logger"
	"matchgame/packet"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

// 處理TCP訊息
func HandleTCPMsg(conn net.Conn, pack packet.Pack) error {
	player := MyRoom.GetPlayerByTCPConn(conn)
	if player == nil {
		log.Errorf("%s HandleMessage 錯誤, 玩家不存在連線清單中", logger.LOG_Room)
		return errors.New("HandleMessage 錯誤, 玩家不存在連線清單中")
	}
	conn.SetDeadline(time.Time{}) // 移除連線超時設定
	// 處理各類型封包
	switch pack.CMD {
	// ==========設定玩家==========
	case packet.SETPLAYER:
		content := packet.SetPlayer{}
		err := json.Unmarshal([]byte(pack.GetContentStr()), &content)
		if err != nil {
			return fmt.Errorf("%s parse %s failed", logger.LOG_Action, pack.CMD)
		}
		// 取mongoDB gladiator doc
		var dbGladiator mongo.DBGladiator
		getDocErr := mongo.GetDocByID(mongo.ColName.Gladiator, content.DBGladiatorID, &dbGladiator)
		if getDocErr != nil {
			return fmt.Errorf("%s 取mongoDB gladiator doc資料發生錯誤: %v", logger.LOG_Action, getDocErr)
		}
		// 設定玩家使用的角鬥士
		player.MyGladiator = &Gladiator{
			ID: dbGladiator.ID,
		}
		log.Infof("%s 收到玩家(%s)的角鬥士(%s)", logger.LOG_Action, player.GetID(), dbGladiator.ID)
		pack := packet.Pack{
			CMD:    packet.SETPLAYER_TOCLIENT,
			PackID: -1,
			Content: &packet.SetPlayer_ToClient{
				Players: MyRoom.GetPackPlayers(),
			},
		}
		log.Infof("BroadCastPacket: %v", pack)
		MyRoom.BroadCastPacket("", pack)
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
