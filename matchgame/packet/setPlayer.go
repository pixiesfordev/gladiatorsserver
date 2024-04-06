package packet

import (
	"gladiatorsGoModule/setting"
)

type SetPlayer struct {
	DBGladiatorID string `json:"DBGladiatorID"`
}

type SetPlayer_ToClient struct {
	Players [setting.PLAYER_NUMBER]PackPlayer
}
type PackPlayer struct {
	DBPlayerID string `json:"DBPlayerID"`
	Gladiator  PackGladiator
}

// func (set *SetPlayer) Parse(content CMDContent) bool {
// 	m := content.(map[string]interface{})

// 	if dbGladiatorID, ok := m["DBGladiatorID"].(string); ok {
// 		set.DBGladiatorID = dbGladiatorID
// 	} else {
// 		log.WithFields(log.Fields{
// 			"log": "parse SpellJsonID資料錯誤",
// 		}).Errorf("%s Parse packet error: %s", logger.LOG_Pack, "Hit")
// 		return false
// 	}

// 	return true
// }
