package setting

import (
	"fmt"
	"gladiatorsGoModule/logger"

	log "github.com/sirupsen/logrus"
)

const (
	// 命名空間
	NAMESPACE_LOBBY      = "gladiators-service"    // 大廳伺服器命名空間
	NAMESPACE_GAMESERVER = "gladiators-gameserver" // 遊戲伺服器命名空間

	// 服務名稱
	LOBBY                 = "gladiators-lobby"         // 大廳伺服器Services名稱
	MATCHGAME_TESTVER_TCP = "gladiators-matchgame-tcp" // 個人測試用遊戲房Services TCP名稱
	MATCHGAME_TESTVER_UDP = "gladiators-matchgame-udp" // 個人測試用遊戲房Services UDP名稱

	// 遊戲房設定
	PLAYER_NUMBER = 2 // 遊戲房最多X位玩家
)

var EnvGCPProject = map[string]string{
	"Dev":     "mygladiators-dev",     // 開發版
	"Release": "mygladiators-release", // 正式版
}

// 登入方式
const (
	AUTH_GUSET  = "GUSET"
	AUTH_GOOGLE = "GOOGLE"
	AUTH_APPLE  = "APPLE"
	AUTH_X      = "X"
	AUTH_WECHAT = "WECHAT"
	AUTH_TIKTOK = "TIKTOK"
)

// 環境版本
const (
	ENV_DEV     = "Dev"
	ENV_TEST    = "Test"
	ENV_RELEASE = "Release"
)

// 專案 MongoDB URI
var mongoURI = map[string]string{
	ENV_DEV:     "mongodb+srv://%s:%s@cluster-gladiators.f9ufimm.mongodb.net/?retryWrites=true&w=majority&appName=cluster-gladiators", // 開發版
	ENV_TEST:    "",                                                                                                                   // 測試用
	ENV_RELEASE: "",                                                                                                                   // 正式版
}

// 專案 MongoDB 名稱
var mongoDBName = map[string]string{
	ENV_DEV:     "gladiators", // 開發版
	ENV_TEST:    "gladiators", // 測試用
	ENV_RELEASE: "gladiators", // 正式版
}

func MongoURI(env, user, pw string) string {

	switch env {
	case ENV_DEV, ENV_TEST, ENV_RELEASE:
		uri := fmt.Sprintf(mongoURI[env], user, pw)
		return uri
	default:
		log.Errorf("%v MongoURI傳入錯誤的env: %v", logger.LOG_Setting, env)
	}
	return ""
}
func MongoDBName(env string) string {
	switch env {
	case ENV_DEV, ENV_TEST, ENV_RELEASE:
		return mongoDBName[env]
	}

	return ""
}
