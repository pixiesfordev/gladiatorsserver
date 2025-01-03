package packet

import (
	"encoding/json"
	// "gladiatorsGoModule/utility"
	"io"
	"strings"

	logger "matchgame/logger"

	log "github.com/sirupsen/logrus"
)

// 封包命令列表
const (
	AUTH                     = "AUTH"                     // (TCP)身分驗證
	AUTH_TOCLIENT            = "AUTH_TOCLIENT"            // (TCP)身分驗證-送Client
	SETPLAYER                = "SETPLAYER"                // (TCP)設定玩家資料
	SETPLAYER_TOCLIENT       = "SETPLAYER_TOCLIENT"       // (TCP)設定玩家資料-送Client
	SETREADY                 = "SETREADY"                 // (TCP)遊戲準備就緒
	SETREADY_TOCLIENT        = "SETREADY_TOCLIENT"        // (TCP)遊戲準備就緒-送Client
	SETDIVINESKILL           = "SETDIVINESKILL"           // (TCP)神祉選擇
	SETDIVINESKILL_TOCLIENT  = "SETDIVINESKILL_TOCLIENT"  // (TCP)神祉選擇-送Client
	PLAYERACTION             = "PLAYERACTION"             // (TCP)玩家指令
	PLAYERACTION_TOCLIENT    = "PLAYERACTION_TOCLIENT"    // (TCP)玩家指令-送Client
	CARDSTATE_TOCLIENT       = "CARDSTATE_TOCLIENT"       // (TCP)手牌更新-送Client
	ENDGAME_TOCLIENT         = "ENDGAME_TOCLIENT"         // (TCP)遊戲結算-送Client
	PING                     = "PING"                     // (TCP)心跳(太久沒收到回傳會視玩家斷線)
	PING_TOCLIENT            = "PING_TOCLIENT"            // (TCP)心跳-送Client(太久沒收到回傳會視玩家斷線)
	GAMESTATE_TOCLIENT       = "GAMESTATE_TOCLIENT"       // (TCP)遊戲狀態-送Client
	MELEE_TOCLIENT           = "MELEE_TOCLIENT"           // (TCP)肉搏-送Client
	BEFORE_MELEE_TOCLIENT    = "BEFORE_MELEE_TOCLIENT"    // (TCP)即將肉搏-送Client
	LOCK_INSTANT_TOCLIENT    = "LOCK_INSTANT_TOCLIENT"    // (TCP)鎖住遠程技能-送Client
	HP_TOCLIENT              = "HP_TOCLIENT"              // (TCP)角鬥士生命-送Client
	KNOCKBACK_TOCLIENT       = "KNOCKBACK_TOCLIENT"       // (TCP)擊退-送Client
	GLADIATORSTATES_TOCLIENT = "GLADIATORSTATES_TOCLIENT" // (TCP)角鬥士狀態-送Client

	GMACTION          = "GMACTION"          // (TCP)GM指令
	GMACTION_TOCLIENT = "GMACTION_TOCLIENT" // (TCP)GM指令-送Client

	UDPAUTH = "UDPAUTH" // (UDP)身分驗證
)

type Pack struct {
	CMD     string
	PackID  int64
	ErrMsg  string
	Content CMDContent
}
type UDPReceivePack struct {
	CMD       string
	PackID    int64
	ErrMsg    string
	ConnToken string // 收到的UPD CMD除了UDPAUTH以外都會包含ConnToken
	Content   CMDContent
}

// 將封包的content內容轉為string
func (pack *Pack) GetContentStr() string {
	// 如果content是interface{}類型那會是map[string]interface{}格式要轉回字串
	if contentMap, ok := pack.Content.(map[string]interface{}); ok {
		contentBytes, err := json.Marshal(contentMap)
		if err != nil {
			return ""
		}
		return string(contentBytes)
	}

	// 如果content是json格式就直接轉字串
	contentStr, ok := pack.Content.(string)
	if ok {
		return contentStr
	}

	// 都不是就返回空字串
	return ""
}

type CMDContent interface {
}

func ReadPack(decoder *json.Decoder) (Pack, error) {
	var packet Pack
	err := decoder.Decode(&packet)

	// 寫LOG
	// log.WithFields(log.Fields{
	// 	"cmd":     packet.CMD,
	// 	"content": packet.Content,
	// 	"error":   packet.ErrMsg,
	// }).Infof("%s Read: %s", logger.LOG_Pack, packet.CMD)
	if err != nil {
		// 檢查是否為EOF錯誤
		if err == io.EOF {
			// 玩家已經斷線，記錄斷線日誌
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Info("目標玩家已斷線-EOF")
		} else if strings.Contains(err.Error(), "connection reset by peer") {
			// 連接被對端重置，記錄斷線日誌
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Info("目標玩家已斷線-connection reset by peer")

		} else if strings.Contains(err.Error(), "use of closed network connection") {
			// 連接被對端重置，記錄斷線日誌
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Info("目標玩家已斷線-use of closed network connection")
		} else {
			// 處理其他類型的錯誤
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("解包packet錯誤")
		}
		return packet, err
	}
	return packet, err
}

func SendPack(encoder *json.Encoder, packet Pack) error {

	err := encoder.Encode(packet)

	if err != nil {
		if strings.Contains(err.Error(), "use of closed network connection") {
			// // 連線已關閉，記錄資訊日誌
			// log.WithFields(log.Fields{
			// 	"error": err.Error(),
			// }).Info("目標玩家已斷線-連線關閉")
		} else {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Errorf("%s 送封包失敗", logger.LOG_Pack)
		}
	}
	return err
}

// 對指定封包進行四捨五入
// func roundDecimalCheck(pack *Pack) (*Pack, error) {
// 	switch pack.CMD {
// 	case BATTLESTATE_TOCLIENT:
// 		if targetPack, ok := pack.Content.(*BattleState_ToClient); ok {
// 			newContent, err := utility.RoundDecimalInStruct(targetPack.MyPlayerState, 3)
// 			if err != nil {
// 				log.Errorf("%s 送封包前將%s封包四捨五入錯誤: %v", logger.LOG_Pack, pack.CMD, err)
// 				return pack, err
// 			} else {
// 				if roundContent, ok := newContent.(PackPlayerState); ok {
// 					targetPack.MyPlayerState = roundContent
// 				} else {
// 					log.Errorf("%s %s utility.RoundDecimalInStruct錯誤: %v", logger.LOG_Pack, pack.CMD, pack.CMD)
// 				}
// 			}

// 			newContent, err = utility.RoundDecimalInStruct(targetPack.OpponentPlayerState, 3)
// 			if err != nil {
// 				log.Errorf("%s 送封包前將%s封包四捨五入錯誤: %v", logger.LOG_Pack, pack.CMD, err)
// 				return pack, err
// 			} else {
// 				if roundContent, ok := newContent.(PackPlayerState); ok {
// 					targetPack.MyPlayerState = roundContent
// 				} else {
// 					log.Errorf("%s %s utility.RoundDecimalInStruct錯誤: %v", logger.LOG_Pack, pack.CMD, pack.CMD)
// 				}
// 			}
// 		}

// 	}
// 	return pack, nil
// }
