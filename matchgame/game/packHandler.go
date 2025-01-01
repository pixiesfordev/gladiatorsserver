package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"gladiatorsGoModule/gameJson"

	// "gladiatorsGoModule/utility"

	// mongo "gladiatorsGoModule/mongo"
	logger "matchgame/logger"
	"matchgame/packet"
	"net"
	"time"

	"github.com/mitchellh/mapstructure"
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

		ChangeGameState(GAMESTATE_WAITINGPLAYERS, true)
		// 設定玩家使用的角鬥士
		gladiator, err := NewTestGladiator(player, 1, []int{1, 1001, 1002, 1003, 1004, 1005}) // 測試用角鬥士
		if err != nil {
			log.Errorf("初始化測試用角鬥士錯誤: %v", err)
		}
		player.MyGladiator = &gladiator

		if Mode == "non-agones" { // 遊戲模式是測試模式時, 自動加入Bot
			AddBot() // 加入BOT
		}

		if MyRoom.GamerCount() == 2 { // 如果雙方都進入房間就設定雙方玩家的Opponent
			opponent := MyRoom.Gamers[0]
			if player == MyRoom.Gamers[0] {
				opponent = MyRoom.Gamers[1]
			}
			player.SetOpponent(opponent)
			player.GetGladiator().Opponent = opponent.GetGladiator()
			opponent.SetOpponent(player)
			opponent.GetGladiator().Opponent = player.GetGladiator()

			if player.MyGladiator != nil && opponent != nil && opponent.GetGladiator() != nil {
				// 回送封包
				myPack := packet.Pack{
					CMD:    packet.SETPLAYER_TOCLIENT,
					PackID: -1,
					Content: &packet.SetPlayer_ToClient{
						Time:               time.Now().UnixMilli(),
						MyPackPlayer:       player.GetPackPlayer(true),
						OpponentPackPlayer: player.GetOpponentPackPlayer(false),
					},
				}
				player.SendPacketToPlayer(myPack)

				// 送對手封包
				opponent, ok := player.GetOpponent().(*Player)
				if ok {
					opponentPack := packet.Pack{
						CMD:    packet.SETPLAYER_TOCLIENT,
						PackID: -1,
						Content: &packet.SetPlayer_ToClient{
							Time:               time.Now().UnixMilli(),
							MyPackPlayer:       opponent.GetPackPlayer(true),
							OpponentPackPlayer: player.GetPackPlayer(false),
						},
					}
					opponent.SendPacketToPlayer(opponentPack)
				}
			}
		}

	// ==========設定準備就緒==========
	case packet.SETREADY:
		player.SetReady()
		playerReadies := MyRoom.GetPlayerReadies()

		if playerReadies[0] && playerReadies[1] {

			// 雙方都準備就就廣播封包
			pack := packet.Pack{
				CMD:    packet.SETREADY_TOCLIENT,
				PackID: -1,
				Content: &packet.SetReady_ToClient{
					PlayerReadies: playerReadies,
				},
			}
			MyRoom.BroadCastPacket(-1, pack)

			// 進入神祉技能倒數
			ChangeGameState(GAMESTATE_SELECTINGDIVINESKILL, true)
			go func() {
				time.Sleep(time.Duration(SelectDivineCountDownSecs) * time.Second) // 等待後進入下一階段
				if MyGameState == GAMESTATE_SELECTINGDIVINESKILL {                 // 如果選神祉倒數結束還沒進入戰鬥開始倒數階段就進入戰鬥開始倒數階段
					ChangeGameState(GAMESTATE_COUNTINGDOWN, true)
					go func() {
						time.Sleep(time.Duration(FightingCountDownSecs * float64(time.Second)))
						StartFighting()
					}()
				}
			}()
		}

	// ==========神祉==========
	case packet.SETDIVINESKILL:
		// 如果已經不是選擇神祇技能階段就返回
		if MyGameState != GAMESTATE_SELECTINGDIVINESKILL {
			return nil
		}

		content := packet.SetDivineSkill{}
		err := json.Unmarshal([]byte(pack.GetContentStr()), &content)
		if err != nil {
			return fmt.Errorf("%s parse %s failed", logger.LOG_Action, pack.CMD)
		}
		var skillIDs [2]int
		for i, jsonID := range content.JsonSkillIDs {
			if jsonID == 0 {
				player.DivineSkills[i] = nil
			} else {
				JsonSkill, err := gameJson.GetJsonSkill(jsonID)
				if err != nil {
					return fmt.Errorf("%s gameJson.GetJsonSkill(jsonID)錯誤: %v", logger.LOG_Action, err)
				}
				player.DivineSkills[i] = &DivineSkill{
					Used:   false,
					MyJson: JsonSkill,
				}
				skillIDs[i] = jsonID
			}
		}
		player.FinishSelectedDivineSkill()

		// 回送封包
		myPack := packet.Pack{
			CMD:    packet.SETDIVINESKILL_TOCLIENT,
			PackID: -1,
			Content: &packet.SetDivineSkill_ToClient{
				DivineSkillIDs: skillIDs,
			},
		}
		player.SendPacketToPlayer(myPack)

		// 如果對手也選好技能就進入戰鬥開始倒數階段
		if player.GetOpponent().IsSelectedDivineSkill() {
			ChangeGameState(GAMESTATE_COUNTINGDOWN, true)
			go func() {
				time.Sleep(time.Duration(FightingCountDownSecs * float64(time.Second))) // 等待後開始戰鬥
				StartFighting()
			}()
		}
	// ==========玩家行為==========
	case packet.PLAYERACTION:
		content := packet.PlayerAction{}
		err := json.Unmarshal([]byte(pack.GetContentStr()), &content)
		if err != nil {
			return fmt.Errorf("%s parse %s failed", logger.LOG_Action, pack.CMD)
		}
		log.Infof("%s 收到玩家動作: %v", logger.LOG_Action, content.ActionType)
		switch content.ActionType {
		case packet.ACTION_SURRENDER: // 投降
			if _, ok := content.ActionContent.(packet.PackAction_Surrender); ok {
				player.Surrender()
				// 回送封包
				pack := packet.Pack{
					CMD:    packet.PLAYERACTION_TOCLIENT,
					PackID: -1,
					Content: &packet.PlayerAction_ToClient{
						PlayerDBID:    player.ID,
						ActionType:    packet.ACTION_SURRENDER,
						ActionContent: &packet.PackAction_Surrender_ToClient{},
					},
				}
				MyRoom.BroadCastPacket(-1, pack)
			} else {
				return fmt.Errorf("%v 轉型錯誤", content.ActionType)
			}
		case packet.ACTION_RUSH: // 衝刺
			log.Infof("%v", content.ActionContent)
			jsonData, err := json.Marshal(content.ActionContent)
			if err != nil {
				log.Errorf("Failed to marshal ActionContent: %v", err)
				return fmt.Errorf("%v  json.Marshal錯誤", content.ActionType)
			}
			var action packet.PackAction_Rush
			err = json.Unmarshal(jsonData, &action)
			if err != nil {
				return fmt.Errorf("%v  json.Unmarshal錯誤", content.ActionType)
			}
			player.GetGladiator().SetRush(action.On)
		case packet.ACTION_SKILL: // 技能施放
			jsonData, err := json.Marshal(content.ActionContent)
			if err != nil {
				log.Errorf("Failed to marshal ActionContent: %v", err)
				player.SendPacketToPlayer_SkillFail() // 送技能施放失敗封包
				return fmt.Errorf("%v  json.Marshal錯誤", content.ActionType)
			}
			var action packet.PackAction_Skill
			err = json.Unmarshal(jsonData, &action)
			if err != nil {
				player.SendPacketToPlayer_SkillFail() // 送技能施放失敗封包
				return fmt.Errorf("%v  json.Unmarshal錯誤", content.ActionType)
			}

			targetSkill, _, err := player.GetGladiator().GetSkill(action.SkillID)
			if err != nil {
				player.SendPacketToPlayer_SkillFail() // 送技能施放失敗封包
				return fmt.Errorf("PackAction_Skill錯誤: %v", err)
			}
			// 施放立即技能 或 啟用肉搏技能
			err = player.GetGladiator().ActiveSkill(targetSkill, action.On)
			if err != nil {
				log.Warnf("ActiveSkill 錯誤: %v", err)
				player.SendPacketToPlayer_SkillFail() // 送技能施放失敗封包
				return err
			}
			// 處理立即技能 與 肉搏技能
			if targetSkill.Activation == gameJson.Instant {
				if action.On {
					myPack := packet.Pack{
						CMD:    packet.PLAYERACTION_TOCLIENT,
						PackID: -1,
						Content: &packet.PlayerAction_ToClient{
							PlayerDBID: player.ID,
							ActionType: packet.INSTANT_SKILL,
							ActionContent: &packet.PackAction_InstantSkill_ToClient{
								SkillID:    targetSkill.ID,
								NewSkilID:  player.GetGladiator().GetHandSkills()[3],
								HandSkills: player.GetGladiator().GetHandSkills(),
							},
						},
					}
					player.SendPacketToPlayer(myPack)

					opponent, ok := player.GetOpponent().(*Player)
					if ok {
						opponentPack := packet.Pack{
							CMD:    packet.PLAYERACTION_TOCLIENT,
							PackID: -1,
							Content: &packet.PlayerAction_ToClient{
								PlayerDBID: player.ID,
								ActionType: packet.INSTANT_SKILL,
								ActionContent: &packet.PackAction_InstantSkill_ToClient{
									SkillID: targetSkill.ID,
								},
							},
						}
						opponent.SendPacketToPlayer(opponentPack)
					}
				}
			} else if targetSkill.Activation == gameJson.Melee {
				myPack := packet.Pack{
					CMD:    packet.PLAYERACTION_TOCLIENT,
					PackID: -1,
					Content: &packet.PlayerAction_ToClient{
						PlayerDBID: player.ID,
						ActionType: packet.ACTIVE_MELEE_SKILL,
						ActionContent: &packet.PackAction_ActiveMeleeSkill_ToClient{
							On:      action.On,
							SkillID: targetSkill.ID,
						},
					},
				}
				player.SendPacketToPlayer(myPack)
			}

		case packet.ACTION_DIVINESKILL: // 神祉技能施放
		default:
			return fmt.Errorf("未定義的ActionType : %s", content.ActionType)
		}
	// ==========GM封包==========
	case packet.GMACTION:
		content := packet.GMAction{}
		err := json.Unmarshal([]byte(pack.GetContentStr()), &content)
		if err != nil {
			return fmt.Errorf("%s parse %s failed", logger.LOG_Action, pack.CMD)
		}
		log.Infof("%s 收到GM動作: %v", logger.LOG_Action, content.ActionType)
		switch content.ActionType {
		case packet.GMACTION_SETGLADIATOR: // 設定角鬥士
			var gmAction packet.PackGMAction_SetGladiator
			err = mapstructure.Decode(content.ActionContent, &gmAction)
			if err != nil {
				return fmt.Errorf("%v封包的Content轉換錯誤: %v", content.ActionType, err)
			}

			// 走跟SetPlayer封包一樣的流程
			ChangeGameState(GAMESTATE_WAITINGPLAYERS, true)
			// 設定玩家使用的角鬥士
			gladiator, err := NewTestGladiator(player, gmAction.GladiatorID, gmAction.SkillIDs[:]) // 測試用角鬥士
			if err != nil {
				log.Errorf("初始化測試用角鬥士錯誤: %v", err)
			}
			player.MyGladiator = &gladiator

			if Mode == "non-agones" { // 遊戲模式是測試模式時, 自動加入Bot
				AddBot() // 加入BOT
			}
			log.Infof("%s 收到GM設定角鬥士 GamerCount(): %v", logger.LOG_Action, MyRoom.GamerCount())
			if MyRoom.GamerCount() == 2 { // 如果雙方都進入房間就設定雙方玩家的Opponent
				opponent := MyRoom.Gamers[0]
				if player == MyRoom.Gamers[0] {
					opponent = MyRoom.Gamers[1]
				}

				if opponent != nil && player.GetGladiator() != nil && opponent.GetGladiator() != nil {
					player.SetOpponent(opponent)
					player.GetGladiator().Opponent = opponent.GetGladiator()
					opponent.SetOpponent(player)
					opponent.GetGladiator().Opponent = player.GetGladiator()

					if player.MyGladiator != nil && opponent != nil && opponent.GetGladiator() != nil {
						// 回送封包
						myPack := packet.Pack{
							CMD:    packet.SETPLAYER_TOCLIENT,
							PackID: -1,
							Content: &packet.SetPlayer_ToClient{
								Time:               time.Now().UnixMilli(),
								MyPackPlayer:       player.GetPackPlayer(true),
								OpponentPackPlayer: player.GetOpponentPackPlayer(false),
							},
						}
						player.SendPacketToPlayer(myPack)

						// 送對手封包
						opponent, ok := player.GetOpponent().(*Player)
						if ok {
							opponentPack := packet.Pack{
								CMD:    packet.SETPLAYER_TOCLIENT,
								PackID: -1,
								Content: &packet.SetPlayer_ToClient{
									Time:               time.Now().UnixMilli(),
									MyPackPlayer:       opponent.GetPackPlayer(true),
									OpponentPackPlayer: player.GetPackPlayer(false),
								},
							}
							opponent.SendPacketToPlayer(opponentPack)
						}
					}

				}

			}
		}
	}

	return nil
}
