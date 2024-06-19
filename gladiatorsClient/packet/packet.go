package packet

import (
	"encoding/json"
	"io"
	"strings"

	logger "gladiatorsClient/logger"

	log "github.com/sirupsen/logrus"
)

// 封包命令列表
const (
	AUTH                  = "AUTH"                 // (TCP)身分驗證
	AUTH_TOCLIENT         = "AUTH_TOCLIENT"        // (TCP)身分驗證-送Client
	SETPLAYER             = "SETPLAYER"            // (TCP)設定玩家資料
	SETPLAYER_TOCLIENT    = "SETPLAYER_TOCLIENT"   // (TCP)設定玩家資料-送Client
	READY                 = "READY"                // (TCP)遊戲準備就緒
	READY_TOCLIENT        = "READY_TOCLIENT"       // (TCP)遊戲準備就緒-送Client
	BRIBE                 = "BRIBE"                // (TCP)神祉選擇
	BRIBE_TOCLIENT        = "BRIBE_TOCLIENT"       // (TCP)神祉選擇-送Client
	PLAYERACTION          = "PLAYERACTION"         // (TCP)玩家指令
	PLAYERACTION_TOCLIENT = "PLAYERACTIONTOCLIENT" // (TCP)玩家指令-送Client
	BATTLESTATE           = "BATTLESTATE"          // (TCP)狀態更新
	BATTLESTATE_TOCLIENT  = "BATTLESTATE_TOCLIENT" // (TCP)狀態更新-送Client
	ENDGAME_TOCLIENT      = "ENDGAME_TOCLIENT"     // (TCP)遊戲結算-送Client
	PING                  = "PING"                 // (TCP)心跳(太久沒收到回傳會視玩家斷線)
	UDPAUTH               = "UDPAUTH"              // (UDP)身分驗證
)

type Pack struct {
	CMD     string
	PackID  int
	ErrMsg  string
	Content CMDContent
}
type UDPReceivePack struct {
	CMD       string
	PackID    int
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

	// // 寫LOG
	// log.WithFields(log.Fields{
	// 	"cmd":     packet.CMD,
	// 	"content": packet.Content,
	// }).Infof("%s Send packet: %s", logger.LOG_Pack, packet.CMD)

	if err != nil {
		// 寫LOG
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorf("%s Send packet encoder error", logger.LOG_Pack)

	}
	return err
}
