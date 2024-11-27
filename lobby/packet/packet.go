package packet

import (
	"encoding/json"
	"io"
	"strings"

	logger "lobby/logger"

	log "github.com/sirupsen/logrus"
)

// 封包命令列表
const (
	AUTH           = "AUTH"           // (TCP)身分驗證
	AUTH_TOCLIENT  = "AUTH_TOCLIENT"  // (TCP)身分驗證-送Client
	PING           = "PING"           // (TCP)心跳(太久沒收到回傳會視玩家斷線)
	PING_TOCLIENT  = "PING_TOCLIENT"  // (TCP)心跳-送Client(太久沒收到回傳會視玩家斷線)
	MATCH          = "MATCH"          // (TCP)配對
	MATCH_TOCLIENT = "MATCH_TOCLIENT" // (TCP)配對-送Client
)

type Pack struct {
	CMD     string
	PackID  int64
	ErrMsg  string
	Content CMDContent
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

	if err != nil {
		// 檢查是否為EOF錯誤
		if err == io.EOF {
			// 玩家已經斷線，記錄斷線日誌
			log.WithFields(log.Fields{
				"error":  err.Error(),
				"cmd":    packet.CMD,
				"packet": packet,
			}).Info("目標玩家已斷線-EOF")
		} else if strings.Contains(err.Error(), "connection reset by peer") ||
			strings.Contains(err.Error(), "use of closed network connection") {
			// 連接被對端重置，記錄斷線日誌
			log.WithFields(log.Fields{
				"error":  err.Error(),
				"cmd":    packet.CMD,
				"packet": packet,
			}).Info("目標玩家已斷線-連線重置")
		} else {
			// 處理其他類型的錯誤
			log.WithFields(log.Fields{
				"error":  err.Error(),
				"cmd":    packet.CMD,
				"packet": packet,
			}).Error("解包packet錯誤")
		}
		return packet, err
	}
	return packet, err
}

func SendPack(encoder *json.Encoder, packet Pack) error {

	err := encoder.Encode(packet)

	if err != nil {
		// 寫LOG
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Errorf("%s Send packet encoder error", logger.LOG_Pack)

	}
	return err
}
