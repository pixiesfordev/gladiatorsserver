package main

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"matchgame/packet"
)

func sendPacket(pack packet.Pack) {
	packetBytes, err := json.Marshal(pack)
	if err != nil {
		log.Infof("封包轉換 JSON 錯誤: %v", err)
		return
	}
	_, err = conn.Write([]byte(string(packetBytes) + "\n"))
	if err != nil {
		log.Infof("發送封包錯誤: %v", err)
	}
}

func send_Auth() {
	pack := packet.Pack{
		CMD: packet.AUTH,
		Content: packet.Auth{
			Token: "",
		},
	}
	sendPacket(pack)
}

func send_Ping() {
	pack := packet.Pack{
		CMD:     packet.PING,
		Content: packet.Ping{},
	}
	sendPacket(pack)
}

func send_UseSkill(skillID int) {
	pack := packet.Pack{
		CMD: packet.PLAYERACTION,
		Content: packet.PlayerAction{
			ActionType: packet.ACTION_SKILL,
			ActionContent: packet.PackAction_Skill{
				On:      true,
				SkillID: skillID,
			},
		},
	}
	sendPacket(pack)
}

func send_SetSkills() {
	pack := packet.Pack{
		CMD: packet.GMACTION,
		Content: packet.GMAction{
			ActionType: packet.GMACTION_SETSKILLS,
			ActionContent: packet.PackGMAction_SetSkills{
				SkillIDs: [6]int{1011, 1012, 1013, 1014, 1015, 1016},
			},
		},
	}
	sendPacket(pack)
}
