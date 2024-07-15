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
			GladiatorPos: GladiatorPos{
				LeftSide:      true,
				CurUnit:       -InitGladiatorPos,
				Speed:         8,
				CantMoveTimer: 0,
			},
			Knockback: 16,
		}

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

		log.Infof("%s 收到玩家(%s)的角鬥士(%s)", logger.LOG_Action, player.GetID(), dbGladiator.ID)
		pack := packet.Pack{
			CMD:    packet.SETPLAYER_TOCLIENT,
			PackID: -1,
			Content: &packet.SetPlayer_ToClient{
				Players: MyRoom.GetPackPlayers(),
			},
		}
		MyRoom.BroadCastPacket(-1, pack)

	// ==========設定準備就緒==========
	case packet.READY:
		player.ready = true
		playerReadies := MyRoom.GetPlayerReadies()
		pack := packet.Pack{
			CMD:    packet.READY_TOCLIENT,
			PackID: -1,
			Content: &packet.Ready_ToClient{
				PlayerReadies: playerReadies,
			},
		}
		MyRoom.BroadCastPacket(-1, pack)

	// ==========神祉==========
	case packet.BRIBE:
		content := packet.Bribe{}
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
			CMD:    packet.BRIBE_TOCLIENT,
			PackID: -1,
			Content: &packet.Bribe_ToClient{
				PlayerStates: MyRoom.GetPackPlayerStates(),
			},
		}
		MyRoom.BroadCastPacket(-1, pack)
		time.Sleep(4 * time.Second)
		GameTime = 0
		MyRoom.ChangeState(GameState_Started)

		// ==========施放技能==========
	case packet.PLAYERACTION:
		log.Infof("%s 收到玩家施放技能: %v", logger.LOG_Action, pack)
		content := packet.PlayerAction{}
		err := json.Unmarshal([]byte(pack.GetContentStr()), &content)
		if err != nil {
			return fmt.Errorf("%s parse %s failed", logger.LOG_Action, pack.CMD)
		}
		switch content.ActionType {
		case packet.PLAYERACTION_RUSH:
			actionRush := packet.PackAction_Rush{}
			actionRushJson, err := json.Marshal(content.ActionContent)
			if err != nil {
				return err
			}
			err = json.Unmarshal(actionRushJson, &actionRush)
			if err != nil {
				return err
			}
			player.GetGladiator().SetRush(actionRush.On, 4)
		default:
			//log.Infof("%s Unknow Player Action: %s, %v", logger.LOG_Player, content.ActionType, content)
			return fmt.Errorf("%s Unknow Player Action: %s, %v", logger.LOG_Pack, content.ActionType, content)
		}
		pStates := [2]packet.PackPlayerState{
			MyRoom.GetPackPlayerStates()[0],
			packet.PackPlayerState{},
		}
		pack := packet.Pack{
			CMD:    packet.PLAYERACTION_TOCLIENT,
			PackID: -1,
			Content: &packet.PlayerAction_ToClient{
				CMDContent:    content,
				ActionType:    content.ActionType,
				ActionContent: content.ActionContent,
				PlayerStates:  [][2]packet.PackPlayerState{pStates},
				GameTime:      sliceMiliSecsToSecs([]int{GameTime}),
			},
		}
		MyRoom.BroadCastPacket(-1, pack)

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
				CMDContent:   content,
				PlayerStates: [][2]packet.PackPlayerState{MyRoom.GetPackPlayerStates()},
				GameTime:     sliceMiliSecsToSecs([]int{GameTime}),
			},
		}
		MyRoom.BroadCastPacket(-1, pack)

	}

	return nil
}
