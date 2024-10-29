package mongo

import (
	"gladiatorsGoModule/setting"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// Collection名稱列表結構
type colStruct struct {
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
	Gladiator     string
	// 遊戲資料
	Matchgame string
}

// Collection名稱列表
var Col = colStruct{

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
	Gladiator:     "gladiator",

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

// BSONToStruct 將 BSON 轉為指定 struct
func BSONToStruct[T any](result interface{}) (*T, error) {
	bsonBytes, err := bson.Marshal(result)
	if err != nil {
		return nil, err
	}

	var output T
	if err := bson.Unmarshal(bsonBytes, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

// DB玩家資料
type DBPlayer struct {
	ID            string            `json:"id" bson:"_id"`
	CreatedAt     time.Time         `json:"createdAt" bson:"createdAt"`
	AuthDatas     map[string]string `json:"authDatas" bson:"authDatas"`
	AuthType      string            `json:"authType" bson:"authType"`
	ConnToken     string            `json:"connToken" bson:"connToken"`
	Gold          int               `json:"gold" bson:"gold"`
	Point         int               `json:"point" bson:"point"`
	OnlineState   string            `json:"onlineState" bson:"onlineState"`
	LastSigninAt  time.Time         `json:"lastSigninAt" bson:"lastSigninAt"`
	LastSignoutAt time.Time         `json:"lastSignoutAt" bson:"lastSignoutAt"`
	Ban           bool              `json:"ban" bson:"ban"`
	DeviceType    string            `json:"deviceType" bson:"deviceType"`
	DeviceUID     string            `json:"deviceUID" bson:"deviceUID"`
	InMatchgameID string            `json:"inMatchgameID" bson:"inMatchgameID"`
	MyGladiatorID string            `json:"myGladiatorID" bson:"myGladiatorID"`
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
	MatchgameTestverIP       string    `bson:"matchgame-testver-tcp-ip"`
	MatchgameTestverPort     int       `bson:"matchgame-testver-port"`
}

// gameSetting的Timer文件
type DBTimer struct {
	ID                  string    `bson:"_id"`
	CreatedAt           time.Time `bson:"createdAt"`
	PlayerOfflineMinute int       `bson:"playerOfflineMinute"`
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
