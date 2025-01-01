package main

import (
	"encoding/json"
	"strings"

	"matchgame/game"
	"matchgame/packet"

	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
)

func processData(data string) {
	var pack packet.Pack
	err := json.Unmarshal([]byte(strings.TrimSpace(data)), &pack)
	if err != nil {
		log.Infof("解析封包錯誤:%v ", err)
		return
	}

	switch pack.CMD {
	case packet.PING_TOCLIENT:
	case packet.AUTH_TOCLIENT:
		log.Infof(">>>>>>>AUTH_TOCLIENT 回應: %v", pack.Content)
		var content packet.Auth_ToClient
		err = mapstructure.Decode(pack.Content, &content)
		if err != nil {
			log.Infof("%v封包的Content轉換錯誤: %v", pack.CMD, err)
			return
		}
		if content.IsAuth {
			log.Infof("帳戶驗證成功")
			toServerPack := packet.Pack{
				CMD: packet.SETPLAYER,
				Content: packet.SetPlayer{
					DBGladiatorID: "660926d4d0b8e0936ddc6afe",
				},
			}
			sendPacket(toServerPack)
		} else {
			log.Infof("帳戶驗證失敗")
		}
	case packet.GAMESTATE_TOCLIENT:
		log.Infof(">>>>>>>GAMESTATE_TOCLIENT 回應: %v", pack.Content)
		var content packet.GameState_ToClient
		err = mapstructure.Decode(pack.Content, &content)
		if err != nil {
			log.Infof("%v封包的Content轉換錯誤: %v", pack.CMD, err)
			return
		}
		changeGameState(game.GameState(content.State))
	case packet.SETPLAYER_TOCLIENT:
		log.Infof(">>>>>>>SETPLAYER_TOCLIENT 回應: %v", pack.Content)
		toServerPack := packet.Pack{
			CMD:     packet.SETREADY,
			Content: packet.SetReady{},
		}
		var content packet.SetPlayer_ToClient
		err = mapstructure.Decode(pack.Content, &content)
		if err != nil {
			log.Infof("%v封包的Content轉換錯誤: %v", pack.CMD, err)
			return
		}
		sm.updateSkills(content.MyPackPlayer.MyPackGladiator.HandSkillIDs[:], 0)
		sendPacket(toServerPack)
	case packet.SETREADY_TOCLIENT:
		log.Infof(">>>>>>>SETREADY_TOCLIENT")
		toServerPack := packet.Pack{
			CMD:    packet.SETDIVINESKILL,
			PackID: 1234, // 替換成適合的封包 ID
			Content: packet.SetDivineSkill{
				JsonSkillIDs: [2]int{0, 0}, // 替換為你的技能 ID
			},
		}
		sendPacket(toServerPack)
		send_SetSkills()
	case packet.SETDIVINESKILL_TOCLIENT:
		// log.Infof(">>>>>>>SETDIVINESKILL_TOCLIENT")
	case packet.PLAYERACTION_TOCLIENT:
		// log.Infof(">>>>>>>PLAYERACTION 回應: %v", pack.Content)
		var content packet.PlayerAction_ToClient
		err = mapstructure.Decode(pack.Content, &content)
		if err != nil {
			log.Infof("%v封包的Content轉換錯誤: %v", pack.CMD, err)
			return
		}
		switch content.ActionType {
		case packet.ACTIVE_MELEE_SKILL:
			var aContent packet.PackAction_ActiveMeleeSkill_ToClient
			err = mapstructure.Decode(content.ActionContent, &aContent)
			if err != nil {
				log.Infof("%v封包的Content轉換錯誤: %v", content.ActionContent, err)
				return
			}
			sm.activeMeleeSkill(aContent.SkillID, aContent.On)
		case packet.INSTANT_SKILL:
			var aContent packet.PackAction_InstantSkill_ToClient
			err = mapstructure.Decode(content.ActionContent, &aContent)
			if err != nil {
				log.Infof("%v封包的Content轉換錯誤: %v", content.ActionContent, err)
				return
			}
			sm.updateSkills(aContent.HandSkills[:], aContent.SkillID)
		}

	case packet.MELEE_TOCLIENT:
		// log.Infof(">>>>>>>MELEE_TOCLIENT")
		var content packet.Melee_ToClient
		err = mapstructure.Decode(pack.Content, &content)
		if err != nil {
			log.Infof("%v封包的Content轉換錯誤: %v", pack.CMD, err)
			return
		}
		if content.MyAttack.SkillID != 0 {
			log.Infof("發動肉搏技能: %v", content.MyAttack.SkillID)
		}
		sm.updateSkills(content.HandSkills[:], content.SkillOnID)

		// log.Infof("手牌: %v 啟用技能: %v", content.MyHandSkillIDs, content.MyAttack.SkillID)
	case packet.HP_TOCLIENT:
		// log.Infof(">>>>>>>HP_TOCLIENT 回應: %v", pack.Content)
		var content packet.Hp_ToClient
		err = mapstructure.Decode(pack.Content, &content)
		if err != nil {
			log.Infof("%v封包的Content轉換錯誤: %v", pack.CMD, err)
			return
		}
		log.Infof("PlayerID: %v  HPChange: %v  ", content.PlayerID, content.HPChange)
	case packet.GLADIATORSTATES_TOCLIENT:
	case packet.GMACTION_TOCLIENT:
	default:
		log.Infof("未定義的 CMD: %v", pack.CMD)
	}
}
