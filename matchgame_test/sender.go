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
	auth := packet.Pack{
		CMD: packet.AUTH,
		Content: packet.Auth{
			Token: "",
		},
	}
	sendPacket(auth)
}

func send_Ping() {
	auth := packet.Pack{
		CMD:     packet.PING,
		Content: packet.Ping{},
	}
	sendPacket(auth)
}

func send_SetSkills() {
	auth := packet.Pack{
		CMD: packet.GMACTION,
		Content: packet.GMAction{
			ActionType: packet.GMACTION_SETSKILLS,
			ActionContent: packet.PackGMAction_SetSkills{
				SkillIDs: [6]int{1011, 1012, 1013, 1014, 1015, 1016},
			},
		},
	}
	sendPacket(auth)
}
