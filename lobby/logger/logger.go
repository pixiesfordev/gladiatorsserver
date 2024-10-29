package logger

import (
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	LOG_Main         = "[Main]"
	LOG_TCP          = "[TCP]"
	LOG_Game         = "[Game]"
	LOG_Agones       = "[Agones]"
	LOG_Room         = "[Room]"
	LOG_Action       = "[Action]"
	LOG_Player       = "[Player]"
	LOG_Bot          = "[Bot]"
	LOG_BotBehaviour = "[BotBehaviour]"
	LOG_Pack         = "[Pake]"
	LOG_Setting      = "[Setting]"
	LOG_GameMath     = "[GameMath]"
	LOG_Handler      = "[LOG_Handler]"
)

func InitLogger() {
	// 設定日誌級別
	log.SetLevel(log.InfoLevel)
	// 設定日誌輸出，預設為標準輸出
	log.SetOutput(os.Stdout)
	// 自定義時間格式，包含毫秒
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
}
