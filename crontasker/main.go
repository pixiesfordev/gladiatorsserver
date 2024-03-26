package main

import (
	logger "crontasker/logger"
	"flag"
	"fmt"
	mongo "gladiatorsGoModule/mongo"
	"os"
	"time"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

// Cron格式參考: https://crontab.cronhub.io/
const (
	PLAYER_OFFLINE_CRON = "*/2 * * * *"  // 玩家離線檢測Cron
	RESET_HEROEXP_CRON  = "*/10 * * * *" // 鬥士經驗重置Cron
)

var Env string // 環境版本

func main() {
	// 設定日誌級別
	log.SetLevel(log.InfoLevel)
	// 設定日誌輸出，預設為標準輸出
	log.SetOutput(os.Stdout)
	// 自定義時間格式，包含毫秒
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	log.Infof("%s ==============Crontasker 啟動==============", logger.LOG_Main)

	// 設定環境版本
	Env = *flag.String("Env", "Dev", "Env setting")
	if envEnv := os.Getenv("Env"); envEnv != "" {
		Env = envEnv
	}
	log.Infof("%s Env: %s", logger.LOG_Main, Env)

	// 初始化MongoDB設定
	mongoAPIPublicKey := os.Getenv("MongoAPIPublicKey")
	mongoAPIPrivateKey := os.Getenv("MongoAPIPrivateKey")
	mongoUser := os.Getenv("MongoUser")
	mongoPW := os.Getenv("MongoPW")
	initMonogo(mongoAPIPublicKey, mongoAPIPrivateKey, mongoUser, mongoPW)

	myCron := cron.New()

	// 新增玩家離線排程
	_, playerOfflineCronErr := myCron.AddFunc(PLAYER_OFFLINE_CRON, playerOfflineHandle)
	if playerOfflineCronErr != nil {
		log.Infof("%s 安排playerOfflineHandler錯誤: %v \n", logger.LOG_Main, playerOfflineCronErr)
		return
	}

	myCron.Start()

	select {}
}

// 初始化MongoDB設定
func initMonogo(mongoAPIPublicKey string, mongoAPIPrivateKey string, user string, pw string) {
	log.Infof("%s 初始化mongo開始", logger.LOG_Main)
	mongo.Init(mongo.InitData{
		Env:           Env,
		APIPublicKey:  mongoAPIPublicKey,
		APIPrivateKey: mongoAPIPrivateKey,
	}, user, pw)
	log.Infof("%s 初始化mongo完成", logger.LOG_Main)
}

// 從player文件中找出olineState為Online的文件_id清單(playerIDs)並從PlayerState表中找_id為playerIDs中且lastUpdatedAt欄位的時間小於minutesBefore的文件
// 如果有表符合以上條件就把對應_id的Player文件的onlineState改為Offline
func playerOfflineHandle() {
	log.Infof("%s 處理玩家離線 \n", logger.LOG_Main)

	// 取Timer設定
	dbTimerDoc := &mongo.DBTimer{}
	err := mongo.GetDocByID(mongo.ColName.GameSetting, "Timer", dbTimerDoc)
	if err != nil {
		fmt.Println("playerOfflineHandler取timer文件錯誤:", err)
		return
	}

	playerIDs, err := mongo.GetDocIDsByFieldValue(mongo.ColName.Player, "onlineState", "Online", mongo.Equal)
	if err != nil {
		fmt.Println("playerOfflineHandler執行mongo.GetDocIDsByFieldValue找Player錯誤:", err)
		return
	}

	if len(playerIDs) <= 0 {
		log.Infof("%s 沒有需要設為離線的玩家 \n", logger.LOG_Main)
		return
	}

	// 計算離線閾值時間
	minutesBefore := time.Now().Add(-time.Duration(dbTimerDoc.PlayerOfflineMinute) * time.Minute)

	// 批量查詢playerState文檔
	var offlinePlayerStates []mongo.DBPlayerState
	err = mongo.GetDocsByFieldValue(mongo.ColName.PlayerState, "_id", playerIDs, mongo.In, &offlinePlayerStates)
	if err != nil {
		fmt.Println("playerOfflineHandler執行mongo.GetDocIDsByFieldValue找PlayerState錯誤:", err)
		return
	}

	// 取需要設為Offline的玩家IDs
	filter := bson.M{
		"$and": []bson.M{
			{"_id": bson.M{"$in": playerIDs}},
			{"lastUpdatedAt": bson.M{"$lt": minutesBefore}},
		},
	}
	offlinePlayerIDs, err := mongo.GetDocIDsByFilter(mongo.ColName.PlayerState, filter)
	if err != nil {
		fmt.Println("查找 playerState 錯誤:", err)
		return
	}

	// 批量更新player文件的onlineState
	if len(offlinePlayerIDs) <= 0 {
		log.Infof("%s 沒有需要設為離線的玩家 \n", logger.LOG_Main)
		return
	}

	updateData := bson.D{{Key: "onlineState", Value: "Offline"}}
	_, updateErr := mongo.UpdateDocsByField(mongo.ColName.Player, "_id", offlinePlayerIDs, updateData)
	if updateErr != nil {
		fmt.Println("批量更新 player onlineState 錯誤:", updateErr)
	}

	log.Infof("%s 處理玩家離線完成, 將%v個玩家設為離線: %v \n", logger.LOG_Main, len(offlinePlayerIDs), offlinePlayerIDs)

}
