package mongo

import (
	"fmt"
	logger "gladiatorsGoModule/logger"

	log "github.com/sirupsen/logrus"
)

var (
	Env           = "Dev" // 目前的環境版本，初始化時會設定
	APIPublicKey  = ""    // 目前的Realm的APIKey，初始化時會設定
	APIPrivateKey = ""    // 目前的Realm的APIKey，初始化時會設定
)

const ()

var EnvDBUri = map[string]string{
	"Dev":     "mongodb+srv://%s:%s@cluster-gladiators.f9ufimm.mongodb.net/?retryWrites=true&w=majority", // 開發版
	"Release": "???",                                                                                     // 正式版
}

var AppEndpoint = map[string]string{
	"Dev":     "https://asia-south1.gcp.data.mongodb-api.com/app/gladiators-pirlo", // 開發版
	"Release": "???",                                                               // 正式版
}

// GroupID就是ProjectID(在atlas app service左上方有垂直三個點那點Project Settings)
// 也可以在開啟Atlas Services時 網址會顯示ProjectID
// 在https://services.cloud.mongodb.com/groups/65b4b62b344719089d82ca3a/apps/65b4c6435b1a5d26443841cc/dashboard中
// https://realm.mongodb.com/groups/[GroupID]/apps/[App ObjectID]/dashboard
var EnvGroupID = map[string]string{
	"Dev":     "65b4b62b344719089d82ca3a", // 開發版
	"Release": "???",                      // 正式版
}

// AppID
var EnvAppID = map[string]string{
	"Dev":     "gladiators-pirlo", // 開發版
	"Release": "???",              // 正式版
}

// App ObjectID跟AppID不一樣, 開啟Atlas Services時 網址會顯示App ObjectID
// https://services.cloud.mongodb.com/groups/65b4b62b344719089d82ca3a/apps/65b4c6435b1a5d26443841cc/dashboard
// https://realm.mongodb.com/groups/[GroupID]/apps/[App ObjectID]/dashboard
var EnvAppObjID = map[string]string{
	"Dev":     "65b4c6435b1a5d26443841cc", // 開發版
	"Release": "???",                      // 正式版
}

var EnvDB = map[string]string{
	"Dev": "gladiators", // 開發版
}

const (
	MATCH_QUICK = "Quick"
)

// 加入玩家
func (dbMatchgame *DBMatchgame) JoinPlayer(playerID string) error {
	// 滿足以下條件之一的房間不可加入
	// 1. 該玩家已在此房間
	// 2. 房間已滿
	joinIdx := -1
	if playerID == "" {
		return fmt.Errorf("要加入的玩家名稱為空")
	}
	for i, v := range dbMatchgame.PlayerIDs {
		if v == playerID {
			return fmt.Errorf("玩家(%s)已經存在DBMatchgame中", playerID)
		}
		if v == "" && joinIdx == -1 {
			joinIdx = i
		}
	}
	if joinIdx == -1 {
		return fmt.Errorf("房間已滿, 玩家(%s)無法加入", playerID)
	}
	dbMatchgame.PlayerIDs[joinIdx] = playerID
	return nil
}

// 移除玩家
func (dbMatchgame *DBMatchgame) KickPlayer(playerID string) {
	for i, v := range dbMatchgame.PlayerIDs {
		if v == playerID {
			dbMatchgame.PlayerIDs[i] = ""
			log.Infof("%s 移除DBMatchgame玩家(%s)", logger.LOG_Mongo, v)
			return
		}
	}
	log.Warnf("%s 移除DBMatchgame玩家(%s)失敗 目標玩家不在清單中", logger.LOG_Mongo, playerID)
}
