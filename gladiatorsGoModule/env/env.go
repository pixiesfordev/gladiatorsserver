package env

import (
	"flag"
	"os"

	"gladiatorsGoModule/logger"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

// 環境版本
const (
	DEV     = "Dev"
	TEST    = "Test"
	RELEASE = "Release"
)

func init() {
	godotenv.Load()
}

// 獲取環境參數:
//
//	v: 參數名稱
//	d: 預設數值
//	comment: flag 備註
//	f: 使用 flag 讀取與否
func GetEnv(v, d, comment string, f bool) string {
	var val string

	// 使用 flag
	if f {
		val = *flag.String(v, d, comment)
	}

	// 如果 .env 有該參數，覆蓋掉原來的值
	if s := os.Getenv(v); s != "" {
		val = s
	}

	if val == "" {
		log.Errorf("%s 取環境變數錯誤 (%s) : 空字串", logger.LOG_Env, v)
	} else {
		log.Infof("%s 取環境變數 (%s) : (%s)", logger.LOG_Env, v, val)
	}

	return val
}
