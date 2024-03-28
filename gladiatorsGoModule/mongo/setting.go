package mongo

import (
	"fmt"
	"gladiatorsGoModule/setting"
	"time"

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
// 在https://realm.mongodb.com/groups/653cd1ccb544ec4945f8df83/apps/653cd937e285e8ddc4d6ac57/dashboard中
// https://realm.mongodb.com/groups/[GroupID]/apps/[App ObjectID]/dashboard
var EnvGroupID = map[string]string{
	"Dev":     "653cd1ccb544ec4945f8df83", // 開發版
	"Release": "???",                      // 正式版
}

// AppID
var EnvAppID = map[string]string{
	"Dev":     "gladiators-pirlo", // 開發版
	"Release": "???",              // 正式版
}

// App ObjectID跟AppID不一樣, 開啟Atlas Services時 網址會顯示App ObjectID
// https://realm.mongodb.com/groups/653cd1ccb544ec4945f8df83/apps/653cd937e285e8ddc4d6ac57/dashboard
// https://realm.mongodb.com/groups/[GroupID]/apps/[App ObjectID]/dashboard
var EnvAppObjID = map[string]string{
	"Dev":     "653cd937e285e8ddc4d6ac57", // 開發版
	"Release": "???",                      // 正式版
}

var EnvDB = map[string]string{
	"Dev": "gladiators", // 開發版
}

const (
	MATCH_QUICK = "Quick"
)

// Collection名稱列表結構
type colNameStruct struct {
	// 遊戲設定
	GameSetting string
	GameLog     string
	Template    string
	Map         string

	// 玩家資料
	Player        string
	PlayerCustom  string
	PlayerState   string
	PlayerHistory string

	// 遊戲資料
	Matchgame string
}

// Collection名稱列表
var ColName = colNameStruct{

	// 遊戲設定
	GameSetting: "gameSetting",
	GameLog:     "gameLog",
	Template:    "template",
	Map:         "map",

	// 玩家資料
	Player:        "player",
	PlayerCustom:  "playerCustom",
	PlayerState:   "playerState",
	PlayerHistory: "playerHistory",

	// 遊戲資料
	Matchgame: "matchgame",
}

type Operator string

const (
	Equal              Operator = "$eq"  // 等於 (Equal) - 指定字段等於給定值
	GreaterThan        Operator = "$gt"  // 大於 (Greater Than) - 指定字段大於給定值
	GreaterThanOrEqual Operator = "$gte" // 大於等於 (Greater Than or Equal) - 指定字段大於或等於給定值
	In                 Operator = "$in"  // 包含於 (In) - 指定字段的值在給定的數組中
	LessThan           Operator = "$lt"  // 小於 (Less Than) - 指定字段小於給定值
	LessThanOrEqual    Operator = "$lte" // 小於等於 (Less Than or Equal) - 指定字段小於或等於給定值
	NotEqual           Operator = "$ne"  // 不等於 (Not Equal) - 指定字段不等於給定值
	NotIn              Operator = "$nin" // 不包含於 (Not In) - 指定字段的值不在給定的數組中
)

// DB玩家資料
type DBPlayer struct {
	ID            string    `bson:"_id"`
	CreatedAt     time.Time `bson:"createdAt"`
	Gold          int       `bson:"gold"`
	Point         int       `bson:"point"`
	Ban           bool      `bson:"ban"`
	InMatchgameID string    `bson:"inMatchgameID"`

	// DB用不到的資料放這
	// AuthType      string    `bson:"authType"`
	// OnlineState   string    `bson:"onlineState"`
	// LastSigninAt  time.Time `bson:"lastSigninAt"`
	// LastSignoutAt time.Time `bson:"lastSignoutAt"`
	// DeviceUID     string    `bson:"deviceUID"`

}

// DB玩家資料
type DBGladiator struct {
	ID              string    `bson:"_id"`
	CreatedAt       time.Time `bson:"createdAt"`
	OwnerID         string    `bson:"ownerID"`
	JsonGladiatorID int       `bson:"jsonGladiatorID"`
	JsonSkillIDs    []int     `bson:"jsonSkillIDs"`
	JsonTraitIDs    []int     `bson:"jsonTraitIDs"`
	JsonEquipIDs    []int     `bson:"jsonEquipIDs"`
	HP              int       `bson:"hp"`
	CurHP           int       `bson:"curHP"`
	VigorRegon      float64   `bson:"vigorRegon"`
	STR             int       `bson:"str"`
	DEF             int       `bson:"def"`
	MDEF            int       `bson:"mdef"`
	CRIT            float64   `bson:"crit"`
	INIT            int       `bson:"init"`
	Knockback       int       `bson:"knockback"`
}

// DB地圖資料
type DBMap struct {
	ID        string    `bson:"_id"`
	CreatedAt time.Time `bson:"createdAt"`
	MatchType string    `bson:"matchType"`
	JsonMapID int       `bson:"jsonMapID"`
	Enable    bool      `bson:"enable"`
}

// gameSetting的GameState文件
type DBGameState struct {
	ID                       string    `bson:"_id"`
	CreatedAt                time.Time `bson:"createdAt"`
	EnvVersion               string    `bson:"envVersion"`
	GameVersion              string    `bson:"gameVersion"`
	MinimumGameVersion       string    `bson:"minimumGameVersion"`
	MatchgameTestverRoomName string    `bson:"matchgame-testver-roomName"`
	MatchgameTestverMapID    string    `bson:"matchgame-testver-mapID"`
	MatchgameTestverIP       string    `bson:"matchgame-testver-ip"`
	MatchgameTestverPort     int       `bson:"matchgame-testver-port"`
}

// gameSetting的Timer文件
type DBTimer struct {
	ID                  string    `bson:"_id"`
	CreatedAt           time.Time `bson:"createdAt"`
	PlayerOfflineMinute int       `bson:"playerOfflineMinute"`
}

// DB玩家狀態資料
type DBPlayerState struct {
	ID           string    `bson:"_id"`
	CreatedAt    time.Time `bson:"createdAt"`
	LastUpdateAt time.Time `bson:"lastUpdatedAt"`
}

// 遊戲房資料
type DBMatchgame struct {
	ID                string                        `bson:"_id"`
	CreatedAt         time.Time                     `bson:"createdAt"`
	DBMapID           string                        `bson:"dbMapID"`
	PlayerIDs         [setting.PLAYER_NUMBER]string `bson:"playerIDs"`
	IP                string                        `bson:"ip"`
	Port              int                           `bson:"port"`
	NodeName          string                        `bson:"nodeName"`
	PodName           string                        `bson:"podName"`
	MatchmakerPodName string                        `bson:"matchmakerPodName"`
}

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
			log.Infof("%s 移除DBMatchgame中 index為%v的玩家(%s)", logger.LOG_Mongo, i, v)
			return
		}
	}
	log.Warnf("%s 移除DBMatchgame玩家(%s)失敗 目標玩家不在清單中", logger.LOG_Mongo, playerID)
}
