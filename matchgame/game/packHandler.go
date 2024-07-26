package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"gladiatorsGoModule/gameJson"

	// mongo "gladiatorsGoModule/mongo"
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
		// var dbGladiator mongo.DBGladiator
		// getDocErr := mongo.GetDocByID(mongo.ColName.Gladiator, content.DBGladiatorID, &dbGladiator)
		// if getDocErr != nil {
		// 	return fmt.Errorf("%s 取mongoDB gladiator doc資料發生錯誤: %v", logger.LOG_Action, getDocErr)
		// }
		// 設定玩家使用的角鬥士
		gladiator, err := NewTestGladiator() // 測試用角鬥士
		player.myGladiator = &gladiator

		if Mode == "non-agones" { // 遊戲模式是測試模式時, 自動加入Bot
			AddBot() // 加入BOT
		}

		// 指定左右先放這。
		if MyRoom.Gamers[0].GetGladiator().LeftSide {
			LeftGamer = MyRoom.Gamers[0]
			RightGamer = MyRoom.Gamers[1]
		} else {
			LeftGamer = MyRoom.Gamers[1]
			RightGamer = MyRoom.Gamers[0]
		}

		// log.Infof("%s 收到玩家(%s)的角鬥士(%s)", logger.LOG_Action, player.GetID(), dbGladiator.ID)
		pack := packet.Pack{
			CMD:    packet.SETPLAYER_TOCLIENT,
			PackID: -1,
			Content: &packet.SetPlayer_ToClient{
				MyPackPlayer:       player.GetPackPlayer(),
				OpponentPackPlayer: player.GetOpponentPackPlayer(),
			},
		}
		MyRoom.BroadCastPacket(-1, pack)

	// ==========設定準備就緒==========
	case packet.READY:
		player.SetReady()
		playerReadies := MyRoom.GetPlayerReadies()
		pack := packet.Pack{
			CMD:    packet.READY_TOCLIENT,
			PackID: -1,
			Content: &packet.Ready_ToClient{
				PlayerReadies: playerReadies,
			},
		}
		MyRoom.BroadCastPacket(-1, pack)
		if playerReadies[0] == true && playerReadies[1] == true {
			ChangeGameState(GameState_SelectingDivineSkill)
		}

	// ==========神祉==========
	case packet.SETDIVINESKILL:
		content := packet.DivineSkill{}
		err := json.Unmarshal([]byte(pack.GetContentStr()), &content)
		if err != nil {
			return fmt.Errorf("%s parse %s failed", logger.LOG_Action, pack.CMD)
		}
		for i, jsonID := range content.JsonSkillIDs {
			if jsonID == 0 {
				player.BribeSkills[i] = nil
			} else {
				JsonSkill, err := gameJson.GetJsonSkill(jsonID)
				if err != nil {
					return fmt.Errorf("%s gameJson.GetJsonSkill(jsonID)錯誤: %v", logger.LOG_Action, err)
				}
				player.BribeSkills[i] = &DivineSkill{
					Used:   false,
					MyJson: JsonSkill,
				}
			}
		}
		pack := packet.Pack{
			CMD:    packet.SETDIVINESKILL_TOCLIENT,
			PackID: -1,
			Content: &packet.DivineSkill_ToClient{
				MyPlayerState:       player.GetPackPlayerState(),
				OpponentPlayerState: player.GetOpponentPackPlayerState(),
			},
		}
		MyRoom.BroadCastPacket(-1, pack)
		time.Sleep(time.Duration(FightingCountDownSecs) * time.Second) // 等待後開始戰鬥
		StartFighting()                                                // 開始戰鬥
		// ==========施放技能==========
	case packet.PLAYERACTION:
		content := packet.PlayerAction{}
		err := json.Unmarshal([]byte(pack.GetContentStr()), &content)
		if err != nil {
			return fmt.Errorf("%s parse %s failed", logger.LOG_Action, pack.CMD)
		}
		log.Infof("%s 收到玩家動作: %v", logger.LOG_Action, content.ActionType)
		switch content.ActionType {
		case packet.PLAYERACTION_RUSH: // 衝刺
			if rushAction, ok := content.ActionContent.(packet.PackAction_Rush); ok {
				player.GetGladiator().SetRush(rushAction.On)
			} else {
				return fmt.Errorf("%s PackAction_Rush轉型錯誤", logger.LOG_Action, pack.CMD)
			}
		case packet.Action_Skill:
			if skillAction, ok := content.ActionContent.(packet.PackAction_Skill); ok {
				player.GetGladiator().ActiveSkill(skillAction.SkillID, skillAction.On)
			} else {
				return fmt.Errorf("%s PackAction_Skill轉型錯誤", logger.LOG_Action, pack.CMD)
			}
		default:
			return fmt.Errorf("未定義的ActionType : %s", content.ActionType)
		}

	case packet.BATTLESTATE:
		content := packet.BattleState{}
		err := json.Unmarshal([]byte(pack.GetContentStr()), &content)
		if err != nil {
			return fmt.Errorf("%s parse %s failed", logger.LOG_Action, pack.CMD)
		}
		pack := packet.Pack{
			CMD:    packet.BATTLESTATE_TOCLIENT,
			PackID: -1,
			Content: &packet.BattleState_ToClient{
				MyPlayerState:       player.GetPackPlayerState(),
				OpponentPlayerState: player.GetOpponentPackPlayerState(),
				GameTime:            GameTime,
			},
		}
		MyRoom.BroadCastPacket(-1, pack)

	}

	return nil
}
