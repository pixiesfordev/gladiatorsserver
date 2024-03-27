package main

import (
	log "github.com/sirupsen/logrus"
	mongo "gladiatorsGoModule/mongo"
	logger "lobby/logger"
	"os"
)

var Env string // 環境版本

// 請求
type RequestData struct {
	Token     string `json:"token"`
	ValueJson string `json:"valueJson"`
}

func main() {
	// 設定日誌級別
	log.SetLevel(log.InfoLevel)
	// 設定日誌輸出，預設為標準輸出
	log.SetOutput(os.Stdout)
	// 自定義時間格式，包含毫秒
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	log.Infof("%s ==============Lobby 啟動==============", logger.LOG_Main)
	Env = os.Getenv("Env")

	// 初始化MongoDB設定
	mongoAPIPublicKey := os.Getenv("MongoAPIPublicKey")
	mongoAPIPrivateKey := os.Getenv("MongoAPIPrivateKey")
	mongoUser := os.Getenv("MongoUser")
	mongoPW := os.Getenv("MongoPW")
	initMonogo(mongoAPIPublicKey, mongoAPIPrivateKey, mongoUser, mongoPW)

	// router := mux.NewRouter()
	// router.HandleFunc("/player/syncredischeck", handleSyncRedisCheck).Methods("POST")
	// router.HandleFunc("/game/getstate", handleGetState).Methods("POST")
	// log.Fatal(http.ListenAndServe(":8080", router))
	log.Infof("%s ==============Lobby 啟動完成==============", logger.LOG_Main)

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
