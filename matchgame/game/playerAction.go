package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"gladiatorsGoModule/gameJson"
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
		player.myGladiator = &Gladiator{
			ID: dbGladiator.ID,
		}
		if Mode == "non-agones" { // 遊戲模式是測試模式時, 自動加入Bot
			AddBot() // 加入BOT
		}

		log.Infof("%s 收到玩家(%s)的角鬥士(%s)", logger.LOG_Action, player.GetID(), dbGladiator.ID)
		pack := packet.Pack{
			CMD:    packet.SETPLAYER_TOCLIENT,
			PackID: -1,
			Content: &packet.SetPlayer_ToClient{
				Players: MyRoom.GetPackPlayers(),
			},
		}
		MyRoom.BroadCastPacket("", pack)
		// ==========設定準備就緒==========
	case packet.READY:
		player.ready = true
		pack := packet.Pack{
			CMD:    packet.READY_TOCLIENT,
			PackID: -1,
			Content: &packet.Ready_ToClient{
				PlayerReadies: MyRoom.GetPlayerReadies(),
			},
		}
		MyRoom.BroadCastPacket("", pack)
	// ==========賄賂==========
	case packet.BRIBE:
		content := packet.Bribe{}
		err := json.Unmarshal([]byte(pack.GetContentStr()), &content)
		if err != nil {
			return fmt.Errorf("%s parse %s failed", logger.LOG_Action, pack.CMD)
		}
		for i, jsonID := range content.JsonBribeIDs {
			if jsonID == 0 {
				player.BribeSkills[i] = nil
			} else {
				jsonBribe, err := gameJson.GetJsonBribe(jsonID)
				if err != nil {
					return fmt.Errorf("%s gameJson.GetJsonBribe(jsonID)錯誤: %v", logger.LOG_Action, err)
				}
				player.BribeSkills[i] = &BribeSkill{
					Used:   false,
					MyJson: jsonBribe,
				}
			}
		}
		pack := packet.Pack{
			CMD:    packet.BRIBE_TOCLIENT,
			PackID: -1,
			Content: &packet.Bribe_ToClient{
				Players:  MyRoom.GetPackPlayerStates(),
				GameTime: GameTime,
			},
		}
		MyRoom.BroadCastPacket("", pack)
	}

	return nil
}
