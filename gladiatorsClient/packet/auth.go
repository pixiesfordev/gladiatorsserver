package packet

import (
	logger "gladiatorsClient/logger"

	log "github.com/sirupsen/logrus"
)

// 帳號登入
type Auth struct {
	CMDContent
	Token string
}

// 帳號登入回傳client
type Auth_ToClient struct {
	CMDContent
	IsAuth    bool   // 是否驗證成功
	ConnToken string // 連線Token
}

func (p *Auth_ToClient) Parse(common CMDContent) bool {
	m := common.(map[string]interface{})
	if value, ok := m["IsAuth"].(bool); ok {
		p.IsAuth = value
	} else {
		// 寫LOG
		log.WithFields(log.Fields{
			"log": "IsAuth資料錯誤",
		}).Errorf("%s Parse packet error: %s", logger.LOG_Pack, "Auth_ToClient")
		return false
	}
	return true
}
